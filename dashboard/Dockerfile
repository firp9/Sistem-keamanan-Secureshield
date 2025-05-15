# Use official Node.js image for build stage
FROM node:16-alpine AS build

# Set working directory
WORKDIR /app

# Copy package files and install dependencies
COPY package.json package-lock.json ./
RUN npm install

# Copy source code and build
COPY . .
RUN npm run build

# Use a lightweight server to serve the build output
FROM node:16-alpine

WORKDIR /app

# Install serve to serve static files
RUN npm install -g serve

# Copy build output from build stage
COPY --from=build /app/build ./build

# Expose port 3000
EXPOSE 3000

# Command to serve the build folder
CMD ["serve", "-s", "build", "-l", "3000"]
