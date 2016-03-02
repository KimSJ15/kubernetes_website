{% if pillar.get('is_systemd') %}
  {% set environment_file = '/etc/sysconfig/docker' %}
{% else %}
  {% set environment_file = '/etc/default/docker' %}
{% endif %}

bridge-utils:
  pkg.installed

{% if grains.os_family == 'RedHat' %}

{{ environment_file }}:
  file.managed:
    - source: salt://docker/default
    - template: jinja
    - user: root
    - group: root
    - mode: 644
    - makedirs: true

{% if (grains.os == 'Fedora' and grains.osrelease_info[0] >= 22) or (grains.os == 'CentOS' and grains.osrelease_info[0] >= 7) %}

docker:
  pkg:
    - installed
  service.running:
    - enable: True
    - require:
      - pkg: docker
    - watch:
      - file: {{ environment_file }}
      - pkg: docker

{% else %}

docker-io:
  pkg:
    - installed

docker:
  service.running:
    - enable: True
    - require:
      - pkg: docker-io
    - watch:
      - file: {{ environment_file }}
      - pkg: docker-io

{% endif %}
{% elif grains.cloud is defined and grains.cloud == 'vsphere' and grains.os == 'Debian' and grains.osrelease_info[0] >=8 %}

{% if pillar.get('is_systemd') %}

{{ pillar.get('systemd_system_path') }}/docker.service:
  file.managed:
    - source: salt://docker/docker.service
    - template: jinja
    - user: root
    - group: root
    - mode: 644
    - defaults:
        environment_file: {{ environment_file }}

# The docker service.running block below doesn't work reliably
# Instead we run our script which e.g. does a systemd daemon-reload
# But we keep the service block below, so it can be used by dependencies
# TODO: Fix this
fix-service-docker:
  cmd.wait:
    - name: /opt/kubernetes/helpers/services bounce docker
    - watch:
      - file: {{ pillar.get('systemd_system_path') }}/docker.service
      - file: {{ environment_file }}
{% endif %}

{{ environment_file }}:
  file.managed:
    - source: salt://docker/docker-defaults
    - template: jinja
    - user: root
    - group: root
    - mode: 644
    - makedirs: true
    - require:
      - pkg: docker-engine

'apt-key':
   cmd.run:
     - name: 'apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D'
     - unless: 'apt-key finger | grep "5811 8E89"'

'apt-update':
  cmd.wait:
    - name: '/usr/bin/apt-get update -y'
    - require:
       - cmd : 'apt-key'

lxc-docker:
  pkg:
    - purged

docker-io:
  pkg:
    - purged

cbr0:
  network.managed:
    - enabled: True
    - type: bridge
    - proto: dhcp
    - ports: none
    - bridge: cbr0
    - delay: 0
    - bypassfirewall: True
    - require_in:
      - service: docker

/etc/apt/sources.list.d/docker.list:
  file.managed:
    - source: salt://docker/docker.list
    - template: jinja
    - user: root
    - group: root
    - mode: 644
    - require:
      - cmd: 'apt-update'

docker-engine:
   pkg:
     - installed
     - require:
       - file: /etc/apt/sources.list.d/docker.list
docker:
   service.running:
     - enable: True
     - require:
       - file: {{ environment_file }}
     - watch:
       - file: {{ environment_file }}

{% else %}

{% if grains.cloud is defined
   and grains.cloud == 'gce' %}
# The default GCE images have ip_forwarding explicitly set to 0.
# Here we take care of commenting that out.
/etc/sysctl.d/11-gce-network-security.conf:
  file.replace:
    - pattern: '^net.ipv4.ip_forward=0'
    - repl: '# net.ipv4.ip_forward=0'

/etc/init.d/docker:
  file.managed:
    - source: salt://docker/docker-init
    - template: jinja
    - user: root
    - group: root
    - mode: 755
    - makedirs: true
{% endif %}

# Work around Salt #18089: https://github.com/saltstack/salt/issues/18089
/etc/sysctl.d/99-salt.conf:
  file.touch

# TODO: This should really be based on network strategy instead of os_family
net.ipv4.ip_forward:
  sysctl.present:
    - value: 1

{{ environment_file }}:
  file.managed:
    - source: salt://docker/docker-defaults
    - template: jinja
    - user: root
    - group: root
    - mode: 644
    - makedirs: true

# Docker is on the ContainerVM image by default. The following
# variables are provided for other cloud providers, and for testing and dire circumstances, to allow
# overriding the Docker version that's in a ContainerVM image.
#
# To change:
#
# 1. Find new deb name at:
#    http://apt.dockerproject.org/repo/pool/main/d/docker-engine
# 2. Download based on that:
#    curl -O http://apt.dockerproject.org/repo/pool/main/d/docker-engine/<deb>
# 3. Upload to GCS:
#    gsutil cp <deb> gs://kubernetes-release/docker/<deb>
# 4. Make it world readable:
#    gsutil acl ch -R -g all:R gs://kubernetes-release/docker/<deb>
# 5. Get a hash of the deb:
#    shasum <deb>
# 6. Update override_deb, override_deb_sha1, override_docker_ver with new
#    deb name, new hash and new version

{% set storage_base='https://storage.googleapis.com/kubernetes-release/docker/' %}

{% set override_deb_url='' %}

{% if grains.get('cloud', '') == 'gce'
   and grains.get('os_family', '') == 'Debian'
   and grains.get('oscodename', '') == 'wheezy' -%}
{% set docker_pkg_name='' %}
{% set override_deb='' %}
{% set override_deb_sha1='' %}
{% set override_docker_ver='' %}

{% elif grains.get('cloud', '') == 'aws'
   and grains.get('os_family', '') == 'Debian'
   and grains.get('oscodename', '') == 'jessie' -%}
# TODO: Get from google storage?
{% set docker_pkg_name='docker-engine' %}
{% set override_docker_ver='1.9.1-0~jessie' %}
{% set override_deb='docker-engine_1.9.1-0~jessie_amd64.deb' %}
{% set override_deb_url='http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.9.1-0~jessie_amd64.deb' %}
{% set override_deb_sha1='c58c39008fd6399177f6b2491222e4438f518d78' %}

# Ubuntu presents as os_family=Debian, osfullname=Ubuntu
{% elif grains.get('cloud', '') == 'aws'
   and grains.get('os_family', '') == 'Debian'
   and grains.get('oscodename', '') == 'trusty' -%}
# TODO: Get from google storage?
{% set docker_pkg_name='docker-engine' %}
{% set override_docker_ver='1.9.1-0~trusty' %}
{% set override_deb='docker-engine_1.9.1-0~trusty_amd64.deb' %}
{% set override_deb_url='http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.9.1-0~trusty_amd64.deb' %}
{% set override_deb_sha1='ce728172ab29f9fdacfffffe2e2f88a144f23875' %}

{% elif grains.get('cloud', '') == 'aws'
   and grains.get('os_family', '') == 'Debian'
   and grains.get('oscodename', '') == 'vivid' -%}
# TODO: Get from google storage?
{% set docker_pkg_name='docker-engine' %}
{% set override_docker_ver='1.9.1-0~vivid' %}
{% set override_deb='docker-engine_1.9.1-0~vivid_amd64.deb' %}
{% set override_deb_url='http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.9.1-0~vivid_amd64.deb' %}
{% set override_deb_sha1='81741f6f16630632de53762c5554238d57b3b9cb' %}

{% elif grains.get('cloud', '') == 'aws'
   and grains.get('os_family', '') == 'Debian'
   and grains.get('oscodename', '') == 'wily' -%}
# TODO: Get from google storage?
{% set docker_pkg_name='docker-engine' %}
{% set override_docker_ver='1.9.1-0~wily' %}
{% set override_deb='docker-engine_1.9.1-0~wily_amd64.deb' %}
{% set override_deb_url='http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.9.1-0~wily_amd64.deb' %}
{% set override_deb_sha1='a505fd49372cf836f5b9ed953053c50b3381dbfd' %}

{% else %}
{% set docker_pkg_name='lxc-docker-1.7.1' %}
{% set override_docker_ver='1.7.1' %}
{% set override_deb='lxc-docker-1.7.1_1.7.1_amd64.deb' %}
{% set override_deb_sha1='81abef31dd2c616883a61f85bfb294d743b1c889' %}
{% endif %}

{% if override_deb_url == '' %}
{% set override_deb_url=storage_base + override_deb %}
{% endif %}

{% if override_docker_ver != '' %}
purge-old-docker-package:
  pkg.removed:
    - pkgs:
      - lxc-docker-1.6.2

/var/cache/docker-install/{{ override_deb }}:
  file.managed:
    - source: {{ override_deb_url }}
    - source_hash: sha1={{ override_deb_sha1 }}
    - user: root
    - group: root
    - mode: 644
    - makedirs: true

# Drop the license file into /usr/share so that everything is crystal clear.
/usr/share/doc/docker/apache.txt:
  file.managed:
    - source: {{ storage_base }}apache2.txt
    - source_hash: sha1=2b8b815229aa8a61e483fb4ba0588b8b6c491890
    - user: root
    - group: root
    - mode: 644
    - makedirs: true

docker-upgrade:
  pkg.installed:
    - sources:
      - {{ docker_pkg_name }}: /var/cache/docker-install/{{ override_deb }}
    - require:
      - file: /var/cache/docker-install/{{ override_deb }}
{% endif %} # end override_docker_ver != ''

# Default docker systemd unit file doesn't use an EnvironmentFile; replace it with one that does.
{% if pillar.get('is_systemd') %}

{{ pillar.get('systemd_system_path') }}/docker.service:
  file.managed:
    - source: salt://docker/docker.service
    - template: jinja
    - user: root
    - group: root
    - mode: 644
    - defaults:
        environment_file: {{ environment_file }}

# The docker service.running block below doesn't work reliably
# Instead we run our script which e.g. does a systemd daemon-reload
# But we keep the service block below, so it can be used by dependencies
# TODO: Fix this
fix-service-docker:
  cmd.wait:
    - name: /opt/kubernetes/helpers/services bounce docker
    - watch:
      - file: {{ pillar.get('systemd_system_path') }}/docker.service
      - file: {{ environment_file }}
{% if override_docker_ver != '' %}
    - require:
      - pkg: docker-upgrade
{% endif %}

{% endif %}

docker:
  service.running:
# Starting Docker is racy on aws for some reason.  To be honest, since Monit
# is managing Docker restart we should probably just delete this whole thing
# but the kubernetes components use salt 'require' to set up a dag, and that
# complicated and scary to unwind.
{% if grains.cloud is defined and grains.cloud == 'aws' %}
    - enable: False
{% else %}
    - enable: True
{% endif %}
    - watch:
      - file: {{ environment_file }}
{% if override_docker_ver != '' %}
      - pkg: docker-upgrade
{% endif %}
{% if pillar.get('is_systemd') %}
      - file: {{ pillar.get('systemd_system_path') }}/docker.service
{% endif %}
{% if override_docker_ver != '' %}
    - require:
      - pkg: docker-upgrade
{% endif %}
{% endif %} # end grains.os_family != 'RedHat'

