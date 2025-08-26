## Drumkit Application

Drumkit is a small end-to-end app that lists loads from Turvo and creates new loads (shipments) in Turvo. It consists of a Go backend and a React + Vite frontend, and an AWS infrastructure defined in Terraform/Terragrunt.

### Quick links and URLs

- Local backend: `http://localhost:8080`
  - Health: `GET /healthz`, `GET /readyz`
  - API: `GET /api/loads`, `POST /api/loads`, `GET /api/loads/{id}`, `GET /api/loads/by-external/{externalTMSLoadID}`, `GET /api/customers`
- Local frontend (Vite): `http://localhost:5173` (proxied to backend for `/api`)

- AWS (workspace-driven domains; see `terraform/drumkit/main.tf`):
  - Service domain pattern: `service.drumkit-<workspace>.<base_domain>` (prod omits `-<workspace>` → `service.drumkit.<base_domain>`)
  - UI domain pattern: `drumkit-<workspace>.<base_domain>` (prod omits `-<workspace>` → `drumkit.<base_domain>`)
  - ALB endpoint (API): value of output `load_balancer_endpoint` from core stack (e.g., `xxx.us-east-1.elb.amazonaws.com`)
  - CloudFront endpoint (UI): output `ui_endpoint` from `terraform/modules/web-app`

Examples:
- Dev API: `https://service.drumkit-dev.<your-domain>` → forwards to ALB → ECS task
- Dev UI: `https://drumkit-dev.<your-domain>` → CloudFront → S3 `-frontend-bucket-blue`

### End-to-end data flow

1. User opens the UI (CloudFront) and the React app loads.
2. The UI calls the backend (`/api/loads`, `/api/customers`, etc.). In local dev, Vite proxies `/api` to `http://localhost:8080`.
3. The Go backend proxies requests to Turvo Public API. It authenticates via OAuth and includes the API key and tenant when configured.
4. Responses from Turvo are mapped into a simplified domain model for the UI.

Sequence for List Loads:
- UI → `GET /api/loads?start&pageSize&...` → Backend handler → Turvo `shipments/list` with whitelisted query filters → Mapper → JSON response with `items` and `pagination`.

Sequence for Create Load:
- UI → `POST /api/loads` with a `Load` payload → Mapper → Turvo `POST /shipments?fullResponse=true` → Mapper → UI.

### Repository layout

- `backend/`: Go service
  - `cmd/server/main.go`: HTTP server entrypoint (chi router, middleware, health, routes)
  - `internal/config`: env + Secrets Manager configuration
  - `internal/http/handlers`: REST handlers (`/api/loads`, `/api/customers`)
  - `internal/turvo`: Turvo client, models, and mapping code
  - `internal/domain`: UI-facing domain types
- `frontend/`: React app (Vite, TypeScript)
  - `src/App.tsx`: grid to list loads
  - `src/components/CreateLoadModal.tsx`: wizard to create a load
- `terraform/`: Infrastructure as code (Terragrunt wrapper)
  - `core`: shared VPC, ALB, ECS cluster resources (consumed by app stack)
  - `drumkit`: app stack (certs, ECS service, web app)
  - `modules`: reusable modules (`ecs-apps`, `web-app`, `certificates`, etc.)

### Backend (Go)

Run locally:
```bash
cd backend
go run ./cmd/server
```

Environment variables (see `backend/internal/config/config.go`):
- `APP_ENV` (default `local`)
- `TURVO_BASE_URL` (default `https://app.turvo.com`)
- `TURVO_API_PREFIX` (default `/v1`)
- OAuth/API: `TURVO_CLIENT_ID`, `TURVO_CLIENT_SECRET`, `TURVO_API_KEY`, `TURVO_USERNAME`, `TURVO_PASSWORD`, `TURVO_SCOPE`, `TURVO_USER_TYPE`, `TURVO_TENANT`
- `ALLOWED_ORIGINS` (CORS origins)
- `TURVO_DEFAULT_CUSTOMER_ID`, `TURVO_DEFAULT_ORIGIN_LOCATION_ID`, `TURVO_DEFAULT_DESTINATION_LOCATION_ID`
- `AWS_REGION`, `SECRETS_MANAGER_TURVO_SECRET_NAME` (optional, when running in AWS)

Key endpoints:
- `GET /healthz` (liveness), `GET /readyz` (readiness)
- `GET /api/loads` (list)
- `POST /api/loads` (create)
- `GET /api/loads/{id}` (get by Turvo shipment id)
- `GET /api/loads/by-external/{externalTMSLoadID}` (find by external id)
- `GET /api/customers` (list minimal customers)

### Frontend (React + Vite)

Run locally:
```bash
cd frontend
pnpm install
pnpm dev
```

Build:
```bash
pnpm build
```

Notes:
- Vite dev server proxies `/api` to the backend at `http://localhost:8080`.
- At build time, you can set `VITE_API_BASE` to an absolute API URL; otherwise the app uses relative `/api` paths.

### Infrastructure (Terraform / Terragrunt)

- App stack: `terraform/drumkit` (certificates, ECS service, Web UI)
  - Stage 1: ACM certs and DNS validation records
  - Stage 3: ECS service behind ALB + S3/CloudFront web app
- Core stack: `terraform/core` (VPC, private subnets, ECS cluster, ALB, etc.)
- See `terraform/drumkit/README.md` for step-by-step deploy (prepare ECR, apply stages, upload UI).

Key outputs and how to use them:
- `acm_dns_records`, `acm_ui_dns_records`: create these CNAMEs in DNS for cert validation
- `service_dns_entry`: CNAME to point your service domain at the ALB
- `ui_dns_entry`: CNAME to point your UI domain at CloudFront

### Local development workflow

Backend:
```bash
cd backend
go run ./cmd/server
```

Frontend:
```bash
cd frontend
pnpm dev
```

Visit `http://localhost:5173`. The UI will call the backend via `/api`.

### API payloads (examples)

Create Load (minimal):
```json
{
  "externalTMSLoadID": "DK-001",
  "status": "NEW",
  "customer": { "name": "Acme" },
  "pickup": { "name": "Acme Warehouse", "addressLine1": "1 Main", "city": "Chicago", "state": "IL", "zipcode": "60601", "country": "US" },
  "consignee": { "name": "Consignee", "addressLine1": "2 Main", "city": "Detroit", "state": "MI", "zipcode": "48201", "country": "US" }
}
```

List Loads (server-side filters forwarded to Turvo):
- `created[gte]`, `updated[lte]`, `status[eq]`, `customId[eq]`, `sortBy`, `start`, `pageSize`, etc.

### Troubleshooting

- Health endpoints: check `GET /healthz` and `GET /readyz` on the service URL.
- Verify ALB target group health checks in AWS (path `/healthz`).
- For the UI, verify CloudFront distribution status and S3 object availability.

