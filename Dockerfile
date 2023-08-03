FROM golang:latest
WORKDIR /hezzl_task_5
COPY . .
RUN go build
CMD ["./hezzl_task_5"]