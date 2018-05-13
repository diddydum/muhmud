FROM node:9 as builder
RUN npm install -g @angular/cli
WORKDIR /src/
COPY . .
RUN npm install
RUN ng build

FROM nginx:1.13
COPY --from=builder /src/dist /usr/share/nginx/html

