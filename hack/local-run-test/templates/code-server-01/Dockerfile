ARG FROM_TAG="latest"

FROM codercom/code-server:${FROM_TAG}

USER 0
RUN apt-get -y update \
&&  apt-get install -y --no-install-recommends \
		wget  \
		vim \
		tmux \
		iputils-ping \
		iproute2 \
		dnsutils \
&& rm -rf /var/lib/apt/lists/*

##---------------------------------------------------
## timezone / locale
##---------------------------------------------------
RUN ln -sf  /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
 && localedef -f UTF-8 -i ja_JP ja_JP.UTF-8
ENV LANG="ja_JP.UTF-8"

##---------------------------------------------------
## kubectl install
##---------------------------------------------------
RUN	KUBE_VER=`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt` \
 && curl -L -o /usr/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/$KUBE_VER/bin/linux/amd64/kubectl" \
 && chmod +x /usr/bin/kubectl

##---------------------------------------------------
## Entry
##---------------------------------------------------
EXPOSE 18080
USER 1000
WORKDIR /home/coder
#ENTRYPOINT ["/bin/sh", "-c", "bash /data/homedir-files/copy-to-homedir.sh && /usr/bin/entrypoint.sh --auth=none --bind-addr 0.0.0.0:18080 ."]
ENTRYPOINT ["/bin/sh", "-c", "/usr/bin/entrypoint.sh --auth=none --bind-addr 0.0.0.0:18080 ."]
