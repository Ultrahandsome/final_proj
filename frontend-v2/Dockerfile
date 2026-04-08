# Build Stage
FROM node:18 AS build-stage

WORKDIR /app

# Copy dependency files only
COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* ./

# Install dependencies (this layer will be cached if package files are unchanged)
RUN yarn install --frozen-lockfile || npm install

# Copy the rest of the application source code (modifying source code won't trigger reinstall)
COPY . .

# Build the project
RUN yarn build || npm run build

# Production Stage
FROM nginx:stable-alpine as production-stage

# Copy the built frontend files
COPY --from=build-stage /app/dist /usr/share/nginx/html

# Copy Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

# Start Nginx server
CMD ["nginx", "-g", "daemon off;"]
