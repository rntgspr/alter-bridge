#!/usr/bin/env bash
# Installs the alter-bridge hooks into Claude Code and Codex.
#
# Claude needs three, because it has no CLI way to be reached from outside:
#   SessionStart      registers the mailbox directories as watch paths
#   UserPromptSubmit  drains pending messages into the turn
#   FileChanged       wakes the session when a message lands
#
# Codex needs one — UserPromptSubmit — because its own CLI carries the nudge
# (`codex queue`), which is what `send` already uses to reach a live session.
#
# Re-running is safe: a group already running one of our commands is reconciled
# in place rather than duplicated, and a file with nothing to change is not
# rewritten at all. Reconciling matters more than it sounds — a matcher that
# drifted (`__from_` instead of `.*__from_.*`) leaves FileChanged silently dead,
# and "already present" would otherwise preserve exactly that. The same goes
# for the bridge path itself: this script gets relocated (e.g. into a plugin
# directory) more often than the subcommand names change, so existing entries
# are matched by subcommand alone and their path is repaired in place — never
# left stale next to a freshly added, correctly pathed duplicate. That match
# covers both the legacy `bash "<script>" <sub>` form and the Go
# `"<bin>" <sub> <provider>` form, so the cutover rewrites bash entries in place.
#
# The hooks run the Go binary `bin/alter-bridge` (build it with `go/build.sh`),
# passing the provider, since the Go entry points take no identity from the
# environment. The bash script next to this installer is kept as a fallback.
#
# This is the only installer either runtime gets — alter-bridge is a plugin on
# both sides, but plugin-bundled hooks aren't a substitute: Codex does not
# execute them yet (openai/codex#16430, open), so both runtimes keep getting
# these hooks written into their global config instead. Run this once after
# cloning/updating the plugin.
#
# Usage: install-hooks.sh [claude|codex]   (no argument installs both)
set -euo pipefail

target="${1:-both}"
case "$target" in
  both|claude|codex) ;;
  *) printf 'install-hooks: target must be claude, codex, or omitted (got "%s")\n' "$target" >&2; exit 2 ;;
esac

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/../../.." && pwd)"
bridge="$repo/bin/alter-bridge"

[ -x "$bridge" ] || { printf 'install-hooks: %s is missing; run go/build.sh first\n' "$bridge" >&2; exit 1; }

# Store the path with $HOME left unexpanded, so the same settings file stays
# valid on another machine or under a different user. The command runs through a
# shell, which expands it.
case "$bridge" in
  "$HOME"/*) bridge_ref="\$HOME${bridge#"$HOME"}" ;;
  *) bridge_ref="$bridge" ;;
esac

python3 - "$bridge_ref" "$target" <<'PY'
import json, os, re, shutil, sys, tempfile, time

bridge, target = sys.argv[1], sys.argv[2]
stamp = time.strftime('%Y%m%d%H%M%S')
home = os.path.expanduser('~')

def short(path):
    return path.replace(home, '~')

def command(sub, provider):
    return '"%s" %s %s' % (bridge, sub, provider)

def wanted_hook(sub, provider, extra):
    h = {'type': 'command', 'command': command(sub, provider), 'timeout': 10}
    h.update(extra or {})
    return h

def find(groups, sub):
    """The group and hook already running this subcommand, if any — matched by
    subcommand alone, since the bridge path drifts across relocations."""
    for group in groups:
        for hook in group.get('hooks', []):
            m = re.match(r'^(?:bash )?"(.*)" (\S+)(?: \S+)?$', hook.get('command', ''))
            if m and m.group(2) == sub:
                return group, hook
    return None, None

# Only what decides *whether* the hook runs is reconciled: the matcher (a stale
# one leaves the hook silently dead) and the type. Cosmetics like timeout and
# additionalContextLimit are set when the entry is created and never forced
# afterwards — on the Codex side a rewrite changes the file's hash and voids its
# trust, so fixing a cosmetic field would cost a disabled hook until the user
# approves it again.
def reconcile(group, hook, matcher, sub, provider):
    """Brings an existing entry back in line. Returns what had drifted."""
    drift = []

    if matcher is None:
        if 'matcher' in group:
            drift.append('matcher %r dropped' % group.pop('matcher'))
    elif group.get('matcher') != matcher:
        drift.append('matcher %r -> %r' % (group.get('matcher'), matcher))
        group['matcher'] = matcher

    if hook.get('type') != 'command':
        drift.append('type %r -> %r' % (hook.get('type'), 'command'))
        hook['type'] = 'command'

    want_cmd = command(sub, provider)
    if hook.get('command') != want_cmd:
        drift.append('command %r -> %r' % (hook.get('command'), want_cmd))
        hook['command'] = want_cmd

    return drift

def write_atomic(path, doc):
    """Never leave a truncated config behind if this dies mid-write."""
    directory = os.path.dirname(path)
    fd, tmp = tempfile.mkstemp(dir=directory, prefix='.install-hooks-', suffix='.json')
    try:
        with os.fdopen(fd, 'w') as f:
            json.dump(doc, f, indent=2)
            f.write('\n')
            f.flush()
            os.fsync(f.fileno())
        os.replace(tmp, path)
    except BaseException:
        if os.path.exists(tmp):
            os.unlink(tmp)
        raise

def apply(path, wanted, seed):
    """Installs or reconciles one settings file. Returns True when it changed."""
    existed = os.path.exists(path)

    if existed:
        with open(path) as f:
            doc = json.load(f)
    else:
        doc = dict(seed)

    hooks = doc.setdefault('hooks', {})
    changed = False

    for event, matcher, sub, provider, extra in wanted:
        groups = hooks.setdefault(event, [])
        group, hook = find(groups, sub)

        if group is None:
            entry = {'hooks': [wanted_hook(sub, provider, extra)]}
            if matcher is not None:
                entry = {'matcher': matcher, **entry}
            groups.append(entry)
            changed = True
            print('  %-18s %-16s ADDED' % (event, sub))
            continue

        drift = reconcile(group, hook, matcher, sub, provider)
        if drift:
            changed = True
            print('  %-18s %-16s FIXED (%s)' % (event, sub, '; '.join(drift)))
        else:
            print('  %-18s %-16s already present' % (event, sub))

    if not changed:
        print('  %s left untouched\n' % short(path))
        return False

    if existed:
        backup = '%s.bak.%s' % (path, stamp)
        shutil.copy2(path, backup)
        print('  backup: %s' % short(backup))
    else:
        os.makedirs(os.path.dirname(path), exist_ok=True)

    write_atomic(path, doc)
    print('  %s written\n' % short(path))
    return True

if target in ('both', 'claude'):
    print('Claude:')
    apply(
        os.path.join(home, '.claude', 'settings.json'),
        [
            ('SessionStart', None, 'watchpaths', 'claude', None),
            ('UserPromptSubmit', None, 'hook', 'claude', None),
            ('FileChanged', '.*__from_.*', 'relay', 'claude', None),
        ],
        seed={},
    )

codex_dir = os.path.join(home, '.codex')
if target == 'claude':
    pass
elif not os.path.isdir(codex_dir):
    print('Codex:\n  ~/.codex absent, skipped')
else:
    print('Codex:')
    # additionalContextLimit stays 0 because the bridge injects whole markdown
    # messages, of no bounded size.
    touched = apply(
        os.path.join(codex_dir, 'hooks.json'),
        [('UserPromptSubmit', None, 'hook', 'codex', {'additionalContextLimit': 0})],
        seed={'description': 'Inject pending alter-bridge messages into Codex prompts.'},
    )
    if touched:
        print('  NOTE: Codex tracks per-hook trust in config.toml under')
        print('        [hooks.state."<path>:user_prompt_submit:0:0"].trusted_hash.')
        print('        A new or changed hook needs approval in Codex before it runs —')
        print('        this file being written does not by itself make it live.')
PY
