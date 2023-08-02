FROM golang:latest
COPY . .
RUN go build
CMD ["./hezzl_task_5"]