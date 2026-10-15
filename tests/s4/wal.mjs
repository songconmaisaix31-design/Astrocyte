import { DatabaseSync } from 'node:sqlite';
import fs from 'node:fs';

// 在真实自有投影上保持一个SQLite连接，验证尚未checkpoint的实际WAL会被查询拒绝。
const databasePath = process.argv[2];
if (!databasePath) throw new Error('owned_projection_path_is_required');
const db = new DatabaseSync(databasePath);
db.exec('BEGIN IMMEDIATE');
const node = db.prepare('SELECT id FROM nodes ORDER BY id LIMIT 1').get();
if (!node) throw new Error('actual_projection_has_no_nodes');
db.prepare('UPDATE nodes SET name = name WHERE id = ?').run(node.id);
db.exec('COMMIT');
if (fs.statSync(`${databasePath}-wal`).size === 0) throw new Error('real_writer_created_no_wal');
process.stdout.write(JSON.stringify({ ready: true }));
process.stdin.resume();
process.stdin.once('end', () => db.close());
