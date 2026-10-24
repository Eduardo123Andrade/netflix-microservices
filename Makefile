.PHONY: run-all run-auth run-users

run-all:
	@$(MAKE) -j2 run-auth run-users

run-auth:
	@$(MAKE) -C servers/auth dev-db-up
	@$(MAKE) -C servers/auth run

run-users:
	@pnpm --dir servers/users dev
