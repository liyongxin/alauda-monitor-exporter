set -e
set -x
#git pull
#git add .
#git commit -m 'default'
#git push origin
docker build .
imageId=$(docker images |awk '{print $3}'|sed -n '2p')
deleteImageId=$(docker images |awk '{print $3}'|sed -n '3p')
echo "image id is $imageId"
docker tag $imageId index.alauda.cn/yxli/alauda-monitor-exporter
docker push index.alauda.cn/yxli/alauda-monitor-exporter
docker rm `docker ps -aq`
docker rmi $imageId $deleteImageId
