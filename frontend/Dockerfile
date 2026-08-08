# --- dev: Vite dev server with hot reload ---
FROM node:22-alpine AS dev
WORKDIR /app
# No COPY: the source arrives as a bind mount, and dependencies install at
# startup into the container's own node_modules volume. The host's are built
# for macOS and must not leak in.
CMD ["sh", "-c", "npm install && npm run dev -- --host 0.0.0.0"]

# --- build the Nuxt app ---
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# --- minimal runtime: just the Nitro server output ---
FROM node:22-alpine AS prod
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/.output ./.output
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
