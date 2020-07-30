build:

clean:

install:
	find mattermost -type d -exec install -Dpm 755 -o 804 -g 804 -d $(DESTDIR)/opt/{} \;
	find mattermost/bin -type f -exec install -Dpm 0755 -o 804 -g 804 {} $(DESTDIR)/opt/{} \;
	for dir in `find mattermost -mindepth 1 -type d ! -name bin`; do find $$dir -type f ! -name config.json -exec install -Dpm 0644 -o 804 -g 804 {} $(DESTDIR)/opt/{} \; ; done
	install -Dpm 644 -t $(DESTDIR)/lib/systemd/system debian/*.service
	install -Dpm 755 -o 804 -g 804 -d $(DESTDIR)/var/log/mattermost

distclean: clean
