#!/bin/bash -eux

# デプロイスクリプト書き換える
APP_NAME=isuride # change
WEBAPP_DIR=/home/isucon/webapp/go # change
SERVICE_NAME=${APP_NAME}-go.service # change

sudo cp -f etc/nginx/nginx.conf /etc/nginx/nginx.conf

# ../${HOSTNAME}/deploy.sh があればそちらを実行して終了
if [ -e ../${HOSTNAME}/deploy.sh ]; then
  ../${HOSTNAME}/deploy.sh
  exit 0
fi

# 各種設定ファイルのコピー
# ../${HOSTNAME}/env.sh があればそちらを優先してコピーする
if [ -e ../${HOSTNAME}/env.sh ]; then
  sudo cp -f ../${HOSTNAME}/env.sh /home/isucon/env.sh
else
  sudo cp -f env.sh /home/isucon/env.sh
fi

# etc以下のファイルについてすべてコピーする
for file in `\find etc -type f`; do
  # .gitkeepはコピーしない
  if [ ${file##*/} = ".gitkeep" ]; then
    continue
  fi

  # 同名のファイルが ../${HOSTNAME}/etc/ にあればそちらを優先してコピーする
  if [ -e ../${HOSTNAME}/$file ]; then
    sudo cp -f ../${HOSTNAME}/$file /$file
    continue
  fi
  sudo cp -f $file /$file
done

# アプリケーションのビルド
cd ${WEBAPP_DIR}

# もしpgo.pb.gzがあればPGOを利用してビルド
if [ -e pgo.pb.gz ]; then
  go build -o ${APP_NAME} -pgo=pgo.pb.gz
else
  go build -o ${APP_NAME}
fi


# ミドルウェア・Appの再起動
sudo systemctl restart mysql
sudo systemctl reload nginx
sudo systemctl restart $SERVICE_NAME

# slow query logの有効化
QUERY="
set global slow_query_log_file = '/var/log/mysql/mysql-slow.log';
set global long_query_time = 0;
set global slow_query_log = ON;
"
echo $QUERY | sudo mysql -uroot

# log permission
sudo chmod -R 777 /var/log/nginx
sudo chmod -R 777 /var/log/mysql
