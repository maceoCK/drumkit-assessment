## Drumkit x Turvo Take-Home

### Overview
Simple web app that lists loads from Turvo and creates a new load (shipment) in Turvo. Backend is Go; frontend is React + Vite.

### Backend (Go)

Env vars:
- `TURVO_BASE_URL` (default `https://app.turvo.com`)
- `TURVO_API_KEY` (provided)
- `TURVO_TENANT` (optional)
- `ALLOWED_ORIGINS` (comma-separated, e.g. `http://localhost:5173`)
- `TURVO_DEFAULT_CUSTOMER_ID` (int, optional)
- `TURVO_DEFAULT_ORIGIN_LOCATION_ID` (int, optional)
- `TURVO_DEFAULT_DESTINATION_LOCATION_ID` (int, optional)
- `APP_ENV` (`local` or `deployed`; `local` loads .env via godotenv)
- `SECRETS_MANAGER_TURVO_SECRET_NAME` (deployed only; JSON with keys `TURVO_API_KEY`, `TURVO_BASE_URL`, `TURVO_TENANT`)

Run locally:
```bash
cd backend
# copy and fill .env (or export envs directly)
cp .env.example .env
# edit .env and set TURVO_API_KEY, etc.
go run ./cmd/server
```

API
- `GET /api/loads` → list loads (mapped from Turvo shipments)
- `POST /api/loads` → create load; minimal example payload:
```json
{
  "externalTMSLoadID": "DK-001",
  "status": "NEW",
  "customer": { "name": "" },
  "pickup": { "name": "", "addressLine1": "", "city": "", "state": "", "zipcode": "", "country": "US" },
  "consignee": { "name": "", "addressLine1": "", "city": "", "state": "", "zipcode": "", "country": "US" },
  "specifications": {}
}
```

### Frontend (React + Vite)

Run locally:
```bash
cd frontend
pnpm install
pnpm dev
```
Vite dev proxy is set to `http://localhost:8080` for `/api`.

### Docker

Build and run backend image:
```bash
cd backend
docker build -t drumkit-backend:local .
docker run --rm -p 8080:8080 \
  -e TURVO_API_KEY=$TURVO_API_KEY \
  -e TURVO_BASE_URL=${TURVO_BASE_URL:-https://app.turvo.com} \
  -e ALLOWED_ORIGINS=http://localhost:5173 \
  drumkit-backend:local
```

### AWS (Console-first)
- Create an ECS Fargate service with a public ALB. Tag resources with `mk`.
- Push the built backend image to ECR, reference in the task definition.
- Configure task env vars for Turvo (see above).
- Allow ALB security group from 0.0.0.0/0 on 80/443.
- Optionally front with CloudFront + WAF.

Frontend can be hosted on S3 + CloudFront; point it at the ALB DNS for API.


