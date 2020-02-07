# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.provider "virtualbox" do |vb|
    vb.memory = "1024"
  end

  config.vm.define "ubuntu" do |node|
    node.vm.box = "ubuntu/xenial64"
    node.vm.synced_folder '../../../', '/home/vagrant/go/src'
    node.vm.provision "shell", inline: <<~SHELL
      apt-get update
      apt-get install -y build-essential xsltproc docbook-xsl libkeepalive0 python quilt \
       devscripts python-setuptools python3 libssl-dev cmake libc-ares-dev uuid-dev daemon \
       sysstat libcurl4-openssl-dev

       sudo add-apt-repository ppa:longsleep/golang-backports
       sudo apt-get update
       sudo apt-get -y install golang-go

       echo "export PATH=$PATH:/usr/local/go/bin" >> /home/vagrant/.bash_profile
       echo "export GOPATH=/home/vagrant/go:$PATH" >> /home/vagrant/.bash_profile
       export GOPATH=/home/vagrant/go
       mkdir -p "$GOPATH/bin"
       chown -R vagrant:vagrant $GOPATH

      if ! [ -x "$(command -v mosquitti)" ]; then
        rm -f mosquitto.tar.gz || true
        curl -L --silent -o mosquitto.tar.gz https://github.com/eclipse/mosquitto/archive/v1.5.8.tar.gz
        rm -rf mosquitto || true
        mkdir -p mosquitto
        tar -xzf mosquitto.tar.gz -C mosquitto
        cd mosquitto/*
        make
        make install
        ldconfig
        cd
      fi

    SHELL
  end
end
