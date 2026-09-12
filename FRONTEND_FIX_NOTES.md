# v1.3.1 local verification fixes

This patch fixes the local verification issues observed on Windows:

- adds Vite `import.meta.env` typing via `src/vite-env.d.ts`
- prevents `tsc -b` from emitting `vite.config.js` / `vite.config.d.ts`
- removes stale generated Vite config artifacts
- separates `useAuth` from `AuthProvider` to satisfy React Fast Refresh linting
- removes redundant `sparse: true` from MongoDB 2dsphere index creation so existing development indexes remain compatible

Re-run from `frontend/`:

```powershell
npm run lint
npm run build
npm run dev
```

Re-run from `backend/`:

```powershell
go test ./...
go run ./cmd/api
```
