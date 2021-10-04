FROM ubuntu:latest

RUN apt update && apt install  openssh-server sudo -y

WORKDIR /tmp
COPY id_rsa.pub id_rsa.pub

RUN useradd -m samir
RUN sudo adduser samir sudo
RUN echo 'samir ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers
RUN su samir && mkdir /home/samir/.ssh && cat /tmp/id_rsa.pub > /home/samir/.ssh/authorized_keys

RUN service ssh start

EXPOSE 22

WORKDIR /home/samir

CMD [ "/usr/sbin/sshd", "-D" ]
