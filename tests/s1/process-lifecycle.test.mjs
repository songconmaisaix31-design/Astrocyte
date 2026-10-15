import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdir, mkdtemp, readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { launchLogged, root, verificationOutputRoot } from '../../scripts/process.mjs';

test('actual child exit drains all stdout and stderr bytes before returning', async () => {
  const output = verificationOutputRoot();
  await mkdir(output, { recursive: true });
  const directory = await mkdtemp(join(output, 'astrocyte-process-drain-'));
  const owned = await launchLogged(process.execPath, [join(root, 'tests/s1/process-output-child.mjs')], { cwd: root }, directory, 'actual-output');
  const exit = await owned.drain();
  assert.equal(exit.code, 0);
  assert.equal(exit.signal, null);
  assert.throws(() => process.kill(owned.child.pid, 0), { code: 'ESRCH' });
  const expected = Buffer.concat(Array(128).fill(await readFile(join(root, 'scripts/process.mjs'))));
  assert.ok(expected.length > 1_000_000, 'actual output must exceed pipe buffering');
  assert.deepEqual(await readFile(owned.log.stdout), expected);
  assert.deepEqual(await readFile(owned.log.stderr), expected);
  assert.deepEqual(await owned.drain(), exit, 'a second drain must retain the same observed exit');
});

test('actual missing executable rejects after retaining both process logs', async () => {
  const output = verificationOutputRoot();
  await mkdir(output, { recursive: true });
  const directory = await mkdtemp(join(output, 'astrocyte-process-unavailable-'));
  const command = join(directory, 'unavailable-native-command');
  const owned = await launchLogged(command, [], { cwd: root }, directory, 'actual-unavailable');
  await assert.rejects(owned.drain(), error => {
    assert.ok(error instanceof AggregateError);
    assert.equal(error.errors[0].code, 'ENOENT');
    return true;
  });
  assert.equal(owned.log.pid, null);
  assert.equal((await readFile(owned.log.stdout)).length, 0);
  assert.equal((await readFile(owned.log.stderr)).length, 0);
  await assert.rejects(launchLogged(process.execPath, ['--version'], { cwd: root }, directory, 'actual-unavailable'), { code: 'EEXIST' });
});
