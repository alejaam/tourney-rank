# Current status

## Verified locally

- Backend: `go test ./...` passes.
- Frontend: `npm run lint` and `npm run build` pass.
- Tournament leaderboard: aggregates verified match reports for registered teams. Scores use `scoring_schema.weights.placement` and `scoring_schema.weights.kills`; either missing weight defaults to `1`.

## Still required before production

- Run the API against a real MongoDB instance and exercise authentication, tournament registration, match moderation, and bracket flows end-to-end.
- Decide and implement the authoritative per-game ranking formula. Player-stat updates currently persist aggregate stats but do not recalculate the global ranking score.
- Implement direct match lookup only if a product screen requires `GET /matches/{id}`; list/history endpoints already cover the current flows.
- Add CI and deployment configuration for the target hosting environment.
