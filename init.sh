#/bin/bash
mount -t tmpfs tmpfs /app
mkdir /app/bin /app/lib /app/proc /app/dev
mount --bind -r /bin /app/bin
mount --bind -r /lib /app/lib
mount -t proc none /app/proc
#mount --bind --make-private /app /tmp/tmp.app
#mkdir /tmp/tmp.app/oldroot.app
#pivot_root /tmp/tmp.app /tmp/tmp.app/oldroot.app

