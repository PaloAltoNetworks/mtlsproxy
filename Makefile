include domingo.mk

codegen: domingo_write_versions
init: domingo_init codegen
test: domingo_test
build: domingo_build domingo_package domingo_package_ca_certificates
