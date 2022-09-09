ARG VERSION=latest
FROM vbatts/slackware:${VERSION}
ARG VERSION=latest
RUN slackpkg update \
  && slackpkg -batch=on -default_answer=y install \
    openssh \
    samba \
    gnutls \
    guile \
    gc \
    make \
    bind \
  && rm -rf /var/lib/slackpkg/*

RUN useradd gvm
RUN echo "gvm:gvm" | chpasswd
RUN mkdir /home/gvm
RUN chown gvm /home/gvm

RUN mkdir -p /usr/local/src/notus/advisories
COPY Makefile /usr/local/src/notus/
COPY ./create_products.sh /usr/local/src/notus/
COPY ./advisories/* /usr/local/src/notus/advisories
WORKDIR /usr/local/src/notus
RUN make create-products
COPY --chmod=600 ${VERSION}_rsa_key /etc/ssh/ssh_host_rsa_key
COPY --chmod=644 ${VERSION}_rsa_key.pub /etc/ssh/ssh_host_rsa_key.pub
COPY --chmod=600 ${VERSION}_dsa_key /etc/ssh/ssh_host_dsa_key
COPY --chmod=644 ${VERSION}_dsa_key.pub /etc/ssh/ssh_host_dsa_key.pub
COPY --chmod=600 ${VERSION}_ecdsa_key /etc/ssh/ssh_host_ecdsa_key
COPY --chmod=644 ${VERSION}_ecdsa_key.pub /etc/ssh/ssh_host_ecdsa_key.pub
COPY --chmod=600 ${VERSION}_ed25519_key /etc/ssh/ssh_host_ed25519_key
COPY --chmod=644 ${VERSION}_ed25519_key.pub /etc/ssh/ssh_host_ed25519_key.pub

CMD /usr/sbin/sshd -D
