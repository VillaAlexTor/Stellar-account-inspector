FROM node:22-alpine AS build
WORKDIR /workspace
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN pnpm install --frozen-lockfile --filter @stellar-inspector/web...
COPY apps/web apps/web
ARG NEXT_PUBLIC_SENTINEL_API_URL=
ENV NEXT_PUBLIC_SENTINEL_API_URL=$NEXT_PUBLIC_SENTINEL_API_URL
RUN pnpm --filter @stellar-inspector/web build

FROM node:22-alpine AS runtime
ENV NODE_ENV=production \
    HOSTNAME=0.0.0.0 \
    PORT=3000
WORKDIR /app/apps/web
COPY --from=build --chown=node:node /workspace/apps/web/.next/standalone /app
COPY --from=build --chown=node:node /workspace/apps/web/.next/static ./.next/static
COPY --from=build --chown=node:node /workspace/apps/web/public ./public
USER node
EXPOSE 3000
CMD ["node", "server.js"]
