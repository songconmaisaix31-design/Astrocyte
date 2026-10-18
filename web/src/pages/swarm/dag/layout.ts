import dagre, { type Point } from '@dagrejs/dagre';
import type { SwarmMissionDAG } from '../../../api/swarm';

export function validateDAG(data: SwarmMissionDAG): SwarmMissionDAG {
  const nodes = new Map(data.nodes.map(node => [node.work_item.id, node]));
  if (nodes.size !== data.nodes.length) throw new Error('任务依赖含重复节点');
  const edges = new Set<string>();
  const graph = new dagre.graphlib.Graph();
  for (const node of data.nodes) {
    if (!node.work_item.id || node.work_item.mission_id !== data.mission_id) throw new Error('任务节点归属无效');
    graph.setNode(node.work_item.id);
  }
  for (const edge of data.edges) {
    const key = JSON.stringify([edge.from, edge.to]);
    if (!nodes.has(edge.from) || !nodes.has(edge.to) || edge.from === edge.to || edges.has(key) || !edge.required) throw new Error('任务依赖含缺失端点、重复边或无效依赖');
    if (!nodes.get(edge.to)?.work_item.dependencies?.includes(edge.from)) throw new Error('任务边缺少实际依赖记录');
    edges.add(key);
    graph.setEdge(edge.from, edge.to);
  }
  for (const node of data.nodes) {
    for (const dependency of node.work_item.dependencies ?? []) {
      if (!edges.has(JSON.stringify([dependency, node.work_item.id]))) throw new Error('任务依赖缺少对应边');
    }
  }
  for (const worker of data.workers) {
    if (!nodes.has(worker.execution.work_item_id)) throw new Error('执行记录引用缺失的工作项节点');
  }
  if (!dagre.graphlib.alg.isAcyclic(graph)) throw new Error('任务依赖存在环');
  return data;
}

function coordinate(value: number | undefined): number {
  if (value === undefined || !Number.isFinite(value)) throw new Error('任务布局返回无效坐标');
  return value;
}

export function layoutDAG(data: SwarmMissionDAG, direction: 'LR' | 'TB') {
  if (data.nodes.length === 0) return { width: 0, height: 0, nodes: [], paths: [] };
  const graph = new dagre.graphlib.Graph();
  graph.setGraph({ rankdir: direction, ranksep: 64, nodesep: 32, marginx: 24, marginy: 24 });
  graph.setDefaultEdgeLabel(() => ({}));
  for (const node of data.nodes) graph.setNode(node.work_item.id, { width: 216, height: 88 });
  for (const edge of data.edges) graph.setEdge(edge.from, edge.to);
  dagre.layout(graph);
  return {
    width: coordinate(graph.graph().width), height: coordinate(graph.graph().height),
    nodes: data.nodes.map(node => ({ ...node, x: coordinate(graph.node(node.work_item.id).x) - 108, y: coordinate(graph.node(node.work_item.id).y) - 44 })),
    paths: data.edges.map(edge => {
      const points: Point[] | undefined = graph.edge(edge.from, edge.to).points;
      if (!points?.length) throw new Error('任务依赖缺少布局路径');
      return { ...edge, path: points.map((point, index) => `${index === 0 ? 'M' : 'L'}${coordinate(point.x)},${coordinate(point.y)}`).join(' ') };
    }),
  };
}
