DEV_MEMBER_NS := toolchain-member-operator
DEV_MEMBER_NS_2 := toolchain-member2-operator
DEV_HOST_NS := toolchain-host-operator
DEV_SSO_NS := toolchain-dev-sso
DEV_REGISTRATION_SERVICE_NS := $(DEV_HOST_NS)
DEV_ENVIRONMENT := dev

SHOW_CLEAN_COMMAND="make clean-dev-resources"

.PHONY: dev-deploy-latest
## Deploy the resources with one member operator instance using the latest and greatest images of Toolchain operators
dev-deploy-latest: DEPLOY_LATEST=true
dev-deploy-latest: dev-deploy-e2e

.PHONY: dev-deploy-e2e
## Deploy the resources with one member operator instance
dev-deploy-e2e: deploy-e2e-to-dev-namespaces print-reg-service-link

.PHONY: dev-deploy-e2e-and-print-local-debug
dev-deploy-e2e-and-print-local-debug: dev-deploy-e2e print-local-debug-info

.PHONY: dev-deploy-e2e-two-members
## Deploy the resources with two instances of member operator
dev-deploy-e2e-two-members: deploy-e2e-to-dev-namespaces-two-members print-reg-service-link

.PHONY: deploy-e2e-to-dev-namespaces
deploy-e2e-to-dev-namespaces:
	$(MAKE) deploy-e2e MEMBER_NS=${DEV_MEMBER_NS} SECOND_MEMBER_MODE=false HOST_NS=${DEV_HOST_NS} REGISTRATION_SERVICE_NS=${DEV_REGISTRATION_SERVICE_NS} ENVIRONMENT=${DEV_ENVIRONMENT} E2E_TEST_EXECUTION=false IS_OSD=${IS_OSD} DEPLOY_LATEST=${DEPLOY_LATEST}
	if [[ ${DEV_SSO} == true ]]; then \
		$(MAKE) download-assets ASSETS_FOLDER=scripts/ci/dev-sso && \
		$(MAKE) run-cicd-script SCRIPT_PATH=scripts/ci/setup-dev-sso.sh  SCRIPT_PARAMS="--sso-ns $(DEV_SSO_NS)"; \
	fi

.PHONY: deploy-e2e-to-dev-namespaces-two-members
deploy-e2e-to-dev-namespaces-two-members:
	$(MAKE) deploy-e2e MEMBER_NS=${DEV_MEMBER_NS} MEMBER_NS_2=${DEV_MEMBER_NS_2} HOST_NS=${DEV_HOST_NS} REGISTRATION_SERVICE_NS=${DEV_REGISTRATION_SERVICE_NS} ENVIRONMENT=${DEV_ENVIRONMENT} E2E_TEST_EXECUTION=false IS_OSD=${IS_OSD}
	if [[ ${DEV_SSO} == true ]]; then \
		$(MAKE) download-assets ASSETS_FOLDER=scripts/ci/dev-sso && \
		$(MAKE) run-cicd-script SCRIPT_PATH=scripts/ci/setup-dev-sso.sh  SCRIPT_PARAMS="--sso-ns $(DEV_SSO_NS)"; \
	fi

.PHONY: dev-deploy-e2e-local
dev-deploy-e2e-local: deploy-e2e-local-to-dev-namespaces print-reg-service-link

.PHONY: dev-deploy-e2e-local-two-members
dev-deploy-e2e-local-two-members: deploy-e2e-local-to-dev-namespaces-two-members print-reg-service-link

.PHONY: deploy-e2e-local-to-dev-namespaces
deploy-e2e-local-to-dev-namespaces:
	$(MAKE) deploy-e2e-local MEMBER_NS=${DEV_MEMBER_NS} SECOND_MEMBER_MODE=false HOST_NS=${DEV_HOST_NS} REGISTRATION_SERVICE_NS=${DEV_REGISTRATION_SERVICE_NS} ENVIRONMENT=${DEV_ENVIRONMENT} E2E_TEST_EXECUTION=false IS_OSD=${IS_OSD}
	if [[ ${DEV_SSO} == true ]]; then \
		$(MAKE) download-assets ASSETS_FOLDER=scripts/ci/dev-sso && \
		$(MAKE) run-cicd-script SCRIPT_PATH=scripts/ci/setup-dev-sso.sh  SCRIPT_PARAMS="--sso-ns $(DEV_SSO_NS)"; \
	fi

.PHONY: deploy-e2e-local-to-dev-namespaces-two-members
deploy-e2e-local-to-dev-namespaces-two-members:
	$(MAKE) deploy-e2e-local MEMBER_NS=${DEV_MEMBER_NS} MEMBER_NS_2=${DEV_MEMBER_NS_2} HOST_NS=${DEV_HOST_NS} REGISTRATION_SERVICE_NS=${DEV_REGISTRATION_SERVICE_NS} ENVIRONMENT=${DEV_ENVIRONMENT} E2E_TEST_EXECUTION=false IS_OSD=${IS_OSD}
	if [[ ${DEV_SSO} == true ]]; then \
		$(MAKE) download-assets ASSETS_FOLDER=scripts/ci/dev-sso && \
		$(MAKE) run-cicd-script SCRIPT_PATH=scripts/ci/setup-dev-sso.sh  SCRIPT_PARAMS="--sso-ns $(DEV_SSO_NS)"; \
	fi


.PHONY: print-reg-service-link
print-reg-service-link:
	@echo ""
	@echo "Deployment complete!"
	@echo "Waiting for the registration-service route being available"
	@echo -n "."
	@while [[ -z `oc get routes registration-service -n ${DEV_REGISTRATION_SERVICE_NS} 2>/dev/null || true` ]]; do \
		if [[ $${NEXT_WAIT_TIME} -eq 100 ]]; then \
            echo ""; \
            echo "The timeout of waiting for the registration-service route has been reached. Try to run 'make  print-reg-service-link' later or check the deployment logs"; \
            exit 1; \
		fi; \
		echo -n "."; \
		sleep 1; \
	done
	@echo ""
	@echo "Waiting for the api route (that is used by proxy) being available"
	@echo -n "."
	@while [[ -z `oc get routes api -n ${DEV_REGISTRATION_SERVICE_NS} 2>/dev/null || true` ]]; do \
		if [[ $${NEXT_WAIT_TIME} -eq 100 ]]; then \
            echo ""; \
            echo "The timeout of waiting for the api route (that is used by proxy) has been reached. Try to run 'make  print-reg-service-link' later or check the deployment logs"; \
            exit 1; \
		fi; \
		echo -n "."; \
		sleep 1; \
	done
	@echo ""
	@echo ""
	@echo "==========================================================================================================================================="
	@echo Access the Landing Page here:   https://$$(oc get routes registration-service -n ${DEV_REGISTRATION_SERVICE_NS} -o=jsonpath='{.spec.host}')
	@echo Access Proxy here:              https://$$(oc get routes api -n ${DEV_REGISTRATION_SERVICE_NS} -o=jsonpath='{.spec.host}')
	@echo "==========================================================================================================================================="
	@echo ""
	@echo "To clean the cluster run '${SHOW_CLEAN_COMMAND}'"
	@echo ""

.PHONY: dev-deploy-e2e-member-local
## Deploy the e2e resources with the local 'member-operator' repository only
dev-deploy-e2e-member-local:
	$(MAKE) dev-deploy-e2e MEMBER_REPO_PATH=${PWD}/../member-operator ENVIRONMENT=${DEV_ENVIRONMENT} E2E_TEST_EXECUTION=false IS_OSD=${IS_OSD}

.PHONY: dev-deploy-e2e-host-local
## Deploy the e2e resource with the local 'host-operator' repository only
dev-deploy-e2e-host-local:
	$(MAKE) dev-deploy-e2e HOST_REPO_PATH=${PWD}/../host-operator ENVIRONMENT=${DEV_ENVIRONMENT} E2E_TEST_EXECUTION=false IS_OSD=${IS_OSD}

.PHONY: dev-deploy-e2e-registration-local
## Deploy the e2e resources with the local 'registration-service' repository only
dev-deploy-e2e-registration-local:
	$(MAKE) dev-deploy-e2e REG_REPO_PATH=${PWD}/../registration-service ENVIRONMENT=${DEV_ENVIRONMENT} E2E_TEST_EXECUTION=false IS_OSD=${IS_OSD}
