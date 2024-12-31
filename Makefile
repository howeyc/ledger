.PHONY: docs clean release snapshot

docs:
	mkdir -p docs
	mandoc -Tpdf -l ledger/man/ledger.1 > docs/ledger.1.pdf
	mandoc -Tpdf -l ledger/man/ledger.5 > docs/ledger.5.pdf
	cp ledger/man/ledger.1 docs/
	cp ledger/man/ledger.5 docs/

snapshot:
	goreleaser --skip-publish --rm-dist --snapshot

release:
	goreleaser

clean:
	rm -rf docs dist

