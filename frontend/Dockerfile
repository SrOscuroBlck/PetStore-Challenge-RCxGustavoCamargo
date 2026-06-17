# Build the SPA (codegen reads the local schema SDL — no network needed for it).
FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Serve the built SPA and act as the same-origin gateway to the API.
FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY deploy/nginx.conf.template /etc/nginx/templates/default.conf.template
# Only substitute our own variable; leave nginx's own $variables untouched.
ENV NGINX_ENVSUBST_FILTER=AMBIENT_AUTH
EXPOSE 80
