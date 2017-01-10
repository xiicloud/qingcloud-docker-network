# 使用方法
1. 登录青云控制台，创建一个VPC网络，然后创建两个私有网络(分别命名为mgmt和user)，并将之加入VPC网络。
2. 基于CentOS 7 64位镜像创建虚拟机，网络选择mgmt。
3. 登录虚拟机，执行 yum remove -y NetworkManager 删除NetworkManger，否则机器上绑定多个网卡时会污染默认路由导致虚拟机无法登录。
4. 向`/etc/sysconfig/network-scripts/ifcfg-eth0`文件里写入以下内容来配置虚拟机主网卡:
  
  ```bash
  TYPE=Ethernet
  DEVICE=eth0
  NAME=eth0
  ONBOOT=yes
  NM_CONTROLLED=no
  BOOTPROTO=dhcp
  ```
5. 执行 `curl https://get.docker.com|bash` 安装Docker
6. 安装并启动青云Docker插件:

  从源码编译
  
  ```bash
  git clone https://github.com/nicescale/qingcloud-docker-network.git /tmp/qingcloud-docker-network
  cd /tmp/qingcloud-docker-network
  make
  cp bin/qingcloud-docker-network /bin/
  ```
  
  或者直接下载编译好的二进制使用：[qingcloud-docker-network.gz](https://github.com/nicescale/qingcloud-docker-network/files/693076/qingcloud-docker-network.gz)
  
  sha1: 3f826f9c5ff13a76bdde11ad08fe5c338dc49ee7
  
  启动：
  
  ```bash
  ACCESS_KEY_ID=xxxxxxxx SECRET_KEY=xxxxxxxxxx ZONE=sh1a /usr/bin/qingcloud-docker-network
  ```

  或者基于Docker镜像运行插件：
  
  ```bash
	docker run -d --restart=always --name=netplugin --net=host \
    --cap-add NET_ADMIN \
    -v /var/lib/docker/qingcloud-network:/var/lib/docker/qingcloud-network \
    -v /var/run/docker/plugins:/var/run/docker/plugins \
    -e ACCESS_KEY_ID=xxxxxxxx \
    -e SECRET_KEY=xxxxxxxxxx \
    -e ZONE=sh1a \
    csphere/qingcloud-docker-network
  
  ```
7. 创建网络:

  ```bash
  docker network create \
    -d qingcloud \
    --subnet=172.25.1.0/24 \
    --gateway=172.25.1.1 \
    -o vxnet=vxnet-qpxj8ci \
    --ipam-driver=qingcloud  \
    --ipam-opt vxnet=vxnet-qpxj8ci \
    vxnet-qpxj8ci

  ```
  
  其中vxnet 填写给容器使用的那个私有网络的ID，可在青云控制台查看。
  gateway和subnet需要在青云把私有网络加入路由器时指定的网络参数一致。
8. 创建容器测试: docker run -it --rm --net=vxnet-qpxj8ci alpine sh

# Copyright and License
Code developed by cSphere (https://csphere.cn) and released under the Apache 2.0 License.

