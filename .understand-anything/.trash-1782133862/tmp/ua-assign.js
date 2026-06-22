const fs = require('fs');
const fn = JSON.parse(fs.readFileSync('/Users/godboutj/code/project_works/.understand-anything/intermediate/arch-filenodes.json','utf8'));

const layers = {
  'layer:api':            {name:'API Layer', desc:'HTTP request handling via go-chi and Huma — route registration, request/response binding, and the OpenAPI interface consumed by the frontend and CLI.', nodes:[]},
  'layer:auth':           {name:'Authentication & Authorization', desc:'Office 365 OIDC login, API keys for automation, sessions, super-admin fallback, and the middleware enforcing group-based project access.', nodes:[]},
  'layer:service':        {name:'Service Layer', desc:'Business logic and orchestration including the git-like versioning subsystem (snapshots, patches, async version worker) that sits between handlers and stores.', nodes:[]},
  'layer:domain':         {name:'Domain & Store Interfaces', desc:'Core domain models (elements, requirements, features, tasks, links, phases, versions) and the pluggable store/cache/file-storage interfaces every other layer depends on.', nodes:[]},
  'layer:persistence':    {name:'Persistence & Storage', desc:'Concrete implementations of the domain store interfaces — PostgreSQL stores, S3 file storage, and the Ristretto in-memory read cache.', nodes:[]},
  'layer:data':           {name:'Data Schema & Migrations', desc:'PostgreSQL schema migrations and the table/schema/endpoint definitions describing the persisted data model and the OpenAPI surface.', nodes:[]},
  'layer:infrastructure': {name:'Infrastructure & CI/CD', desc:'Docker images, Docker Compose topology (backend, database, file storage), test container, and GitHub Actions packaging/test pipelines.', nodes:[]},
  'layer:config':         {name:'Configuration & Tooling', desc:'Application config loading, the config.yaml / go.mod project config, and build/dev tooling scripts (justfile, test results parser, entry points).', nodes:[]},
  'layer:documentation':  {name:'Documentation', desc:'Project documentation: README, ARCHITECTURE, DATA_MODEL, backend DIAGRAM, and Claude guidance.', nodes:[]},
};

function assign(n){
  const id=n.id, p=n.filePath||'', t=n.type;
  // Non-code by type first
  if(t==='document') return 'layer:documentation';
  if(t==='pipeline') return 'layer:infrastructure';
  if(t==='service') return 'layer:infrastructure';
  if(t==='table') return 'layer:data';
  if(t==='schema') return 'layer:data';   // sql migrations + openapi schema
  if(t==='endpoint') return 'layer:data'; // openapi endpoint definitions
  if(t==='config') return 'layer:config';
  // code files by path
  if(/\/internal\/api\//.test(p)) return 'layer:api';
  if(/\/internal\/auth\//.test(p)) return 'layer:auth';
  if(/\/internal\/service\//.test(p)) return 'layer:service';
  if(/\/internal\/domain\//.test(p)) return 'layer:domain';
  if(/\/internal\/store\//.test(p)) return 'layer:persistence';
  if(/\/internal\/filestore\//.test(p)) return 'layer:persistence';
  if(/\/internal\/cache\//.test(p)) return 'layer:persistence';
  if(/\/internal\/config\//.test(p)) return 'layer:config';
  if(/\/cmd\//.test(p)) return 'layer:config'; // entry points + integration test harness
  if(/justfile$|TestsResultsParser\.sh$/.test(p)) return 'layer:config';
  return 'layer:config';
}

for(const n of fn){
  const l=assign(n);
  layers[l].nodes.push(n.id);
}

const out=[];
for(const [id,v] of Object.entries(layers)){
  if(v.nodes.length===0) continue;
  out.push({id, name:v.name, description:v.desc, nodeIds:v.nodes});
}
const total=out.reduce((s,l)=>s+l.nodeIds.length,0);
fs.writeFileSync('/Users/godboutj/code/project_works/.understand-anything/intermediate/layers.json', JSON.stringify(out,null,2));
console.log('layers:',out.length,'total nodes:',total,'input nodes:',fn.length);
for(const l of out) console.log(' ',l.id, l.nodeIds.length);
