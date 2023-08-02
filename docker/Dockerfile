FROM golang:latest


WORKDIR ../build

COPY . .

RUN go build -o hezzl_task_5 ../server/cmd

CMD ["./hezzl_task_5"]