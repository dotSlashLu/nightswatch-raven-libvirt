CC=go

libvirt.so:
	$(CC) build --buildmode=plugin libvirt.go

install: libvirt.so
	mkdir -p /var/lib/nwatch/
	mv libvirt.so /var/lib/nwatch/

.PHONY: clean

clean:
	rm -f libvirt.so
