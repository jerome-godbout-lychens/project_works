#!/usr/bin/env node
'use strict';
const fs = require('fs');

function main() {
  const inPath = process.argv[2];
  const outPath = process.argv[3];
  if (!inPath || !outPath) { console.error('usage: analyze.js <in> <out>'); process.exit(1); }

  const data = JSON.parse(fs.readFileSync(inPath, 'utf8'));
  const nodes = data.nodes || [];
  const edges = data.edges || [];
  const layers = data.layers || [];

  const byId = new Map();
  for (const n of nodes) byId.set(n.id, n);

  // Fan-in / fan-out
  const fanIn = new Map();
  const fanOut = new Map();
  for (const n of nodes) { fanIn.set(n.id, 0); fanOut.set(n.id, 0); }
  // adjacency for imports/calls
  const fwd = new Map(); // source -> [targets] for imports/calls
  for (const n of nodes) fwd.set(n.id, []);
  for (const e of edges) {
    if (byId.has(e.source) && byId.has(e.target)) {
      fanOut.set(e.source, fanOut.get(e.source) + 1);
      fanIn.set(e.target, fanIn.get(e.target) + 1);
      if (e.type === 'imports' || e.type === 'calls') {
        fwd.get(e.source).push(e.target);
      }
    }
  }

  const nm = id => (byId.get(id) || {}).name || id;
  const sm = id => (byId.get(id) || {}).summary || '';

  const fanInRanking = [...fanIn.entries()]
    .map(([id, c]) => ({ id, fanIn: c, name: nm(id) }))
    .sort((a, b) => b.fanIn - a.fanIn).slice(0, 20);
  const fanOutRanking = [...fanOut.entries()]
    .map(([id, c]) => ({ id, fanOut: c, name: nm(id) }))
    .sort((a, b) => b.fanOut - a.fanOut).slice(0, 20);

  // Entry point scoring
  const codeEntryNames = new Set(['index.ts','index.js','main.ts','main.js','app.ts','app.js','server.ts','server.js','mod.rs','main.go','main.py','main.rs','manage.py','app.py','wsgi.py','asgi.py','run.py','__main__.py','Application.java','Main.java','Program.cs','config.ru','index.php','App.swift','Application.kt','main.cpp','main.c']);
  const fanOutVals = [...fanOut.values()].sort((a,b)=>a-b);
  const fanInVals = [...fanIn.values()].sort((a,b)=>a-b);
  const top10pctFanOut = fanOutVals[Math.floor(fanOutVals.length*0.9)] || 0;
  const bottom25FanIn = fanInVals[Math.floor(fanInVals.length*0.25)] || 0;

  const epScores = [];
  for (const n of nodes) {
    let score = 0;
    const fp = n.filePath || '';
    const depth = fp.split('/').length;
    if (n.type === 'document') {
      if (n.name === 'README.md' && depth === 1) score += 5;
      else if (/\.md$/.test(n.name) && depth === 1) score += 2;
    } else if (n.type === 'file') {
      if (codeEntryNames.has(n.name)) score += 3;
      if (depth <= 2) score += 1;
      if (fanOut.get(n.id) >= top10pctFanOut && top10pctFanOut > 0) score += 1;
      if (fanIn.get(n.id) <= bottom25FanIn) score += 1;
    }
    if (score > 0) epScores.push({ id: n.id, score, name: n.name, summary: sm(n.id), type: n.type });
  }
  epScores.sort((a, b) => b.score - a.score);
  const entryPointCandidates = epScores.slice(0, 5);

  // BFS from top code entry point
  const codeEP = epScores.find(e => e.type === 'file');
  const startNode = codeEP ? codeEP.id : (nodes.find(n=>n.type==='file')||{}).id;
  const order = [];
  const depthMap = {};
  if (startNode) {
    const q = [[startNode, 0]];
    depthMap[startNode] = 0;
    const seen = new Set([startNode]);
    while (q.length) {
      const [cur, d] = q.shift();
      order.push(cur);
      for (const t of (fwd.get(cur) || [])) {
        if (!seen.has(t)) { seen.add(t); depthMap[t] = d + 1; q.push([t, d + 1]); }
      }
    }
  }
  const byDepth = {};
  for (const [id, d] of Object.entries(depthMap)) {
    (byDepth[d] = byDepth[d] || []).push(id);
  }

  // Non-code inventory
  const nonCodeFiles = { documentation: [], infrastructure: [], data: [], config: [] };
  for (const n of nodes) {
    const item = { id: n.id, name: n.name, type: n.type, summary: sm(n.id) };
    if (n.type === 'document') nonCodeFiles.documentation.push(item);
    else if (['service','pipeline','resource'].includes(n.type)) nonCodeFiles.infrastructure.push(item);
    else if (['table','schema','endpoint'].includes(n.type)) nonCodeFiles.data.push(item);
    else if (n.type === 'config') nonCodeFiles.config.push(item);
  }

  // Clusters: bidirectional imports/calls
  const pairKey = (a,b) => [a,b].sort().join('|');
  const directed = new Set();
  for (const e of edges) {
    if ((e.type === 'imports' || e.type === 'calls') && byId.has(e.source) && byId.has(e.target)) {
      directed.add(e.source + '>>' + e.target);
    }
  }
  const clustersMap = new Map();
  for (const e of edges) {
    if ((e.type === 'imports' || e.type === 'calls') && directed.has(e.target + '>>' + e.source)) {
      const k = pairKey(e.source, e.target);
      if (!clustersMap.has(k)) clustersMap.set(k, new Set([e.source, e.target]));
    }
  }
  // edge count between cluster members
  const edgeCountBetween = (set) => {
    let c = 0;
    for (const e of edges) if (set.has(e.source) && set.has(e.target)) c++;
    return c;
  };
  // expand: add nodes connecting to 2+ members
  const adjAll = new Map();
  for (const n of nodes) adjAll.set(n.id, new Set());
  for (const e of edges) {
    if (byId.has(e.source) && byId.has(e.target) && (e.type==='imports'||e.type==='calls')) {
      adjAll.get(e.source).add(e.target);
      adjAll.get(e.target).add(e.source);
    }
  }
  const clusters = [];
  for (const set of clustersMap.values()) {
    let changed = true;
    while (changed && set.size < 5) {
      changed = false;
      for (const n of nodes) {
        if (set.has(n.id)) continue;
        let conn = 0;
        for (const m of set) if (adjAll.get(n.id).has(m)) conn++;
        if (conn >= 2) { set.add(n.id); changed = true; if (set.size >= 5) break; }
      }
    }
    clusters.push({ nodes: [...set], edgeCount: edgeCountBetween(set) });
  }
  // dedupe & sort
  const seenC = new Set();
  const uniqClusters = [];
  clusters.sort((a,b)=>b.edgeCount-a.edgeCount);
  for (const c of clusters) {
    const k = [...c.nodes].sort().join('|');
    if (seenC.has(k)) continue; seenC.add(k); uniqClusters.push(c);
  }
  const topClusters = uniqClusters.slice(0, 10);

  // node summary index
  const nodeSummaryIndex = {};
  for (const n of nodes) nodeSummaryIndex[n.id] = { name: n.name, type: n.type, summary: n.summary || '' };

  const out = {
    scriptCompleted: true,
    entryPointCandidates,
    fanInRanking,
    fanOutRanking,
    bfsTraversal: { startNode, order, depthMap, byDepth },
    nonCodeFiles,
    clusters: topClusters,
    layers: { count: layers.length, list: layers.map(l => ({ id: l.id, name: l.name, description: l.description })) },
    nodeSummaryIndex,
    totalNodes: nodes.length,
    totalEdges: edges.length
  };
  fs.writeFileSync(outPath, JSON.stringify(out, null, 2));
  console.error('done. entry=' + startNode + ' bfsReached=' + order.length);
}
try { main(); } catch (e) { console.error(e.stack || String(e)); process.exit(1); }
