package parallel

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/codeready-toolchain/toolchain-common/pkg/cluster"
	commonsocialevent "github.com/codeready-toolchain/toolchain-common/pkg/socialevent"
	"github.com/codeready-toolchain/toolchain-common/pkg/states"
	commonauth "github.com/codeready-toolchain/toolchain-common/pkg/test/auth"
	testsocialevent "github.com/codeready-toolchain/toolchain-common/pkg/test/socialevent"
	. "github.com/codeready-toolchain/toolchain-e2e/testsupport"
	authsupport "github.com/codeready-toolchain/toolchain-e2e/testsupport/auth"
	"github.com/codeready-toolchain/toolchain-e2e/testsupport/cleanup"
	"github.com/codeready-toolchain/toolchain-e2e/testsupport/wait"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func TestLandingPageReachable(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	// when & then
	// just make sure that the landing page is reachable
	wait.NewHTTPRequest().Method("GET").
		URL(route).
		RequireStatusCode(http.StatusOK).
		Execute(t)
}

func TestHealth(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	t.Run("get healthcheck 200 OK", func(t *testing.T) {

		// when
		// Call health endpoint.
		mp, _ := wait.NewHTTPRequest().
			Method("GET").
			URL(route + "/api/v1/health").
			RequireStatusCode(http.StatusOK).
			Execute(t)

		// then
		// Verify JSON response.
		alive := mp["alive"]
		require.IsType(t, true, alive)
		require.True(t, alive.(bool))

		environment := mp["environment"]
		require.IsType(t, "", environment)
		require.Equal(t, "e2e-tests", environment.(string))

		revision := mp["revision"]
		require.NotNil(t, revision)

		buildTime := mp["buildTime"]
		require.NotNil(t, buildTime)

		startTime := mp["startTime"]
		require.NotNil(t, startTime)
	})
}

func TestWoopra(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	assertNotSecuredGetResponseEquals := func(endPointPath, expectedResponseValue string) {
		// when & then
		// Call woopra domain endpoint.
		wait.NewHTTPRequest().
			Method("GET").
			URL(fmt.Sprintf("%s/api/v1/%s", route, endPointPath)).
			RequireStatusCode(http.StatusOK).
			RequireResponseBody(expectedResponseValue).
			Execute(t)
	}

	t.Run("get woopra domain 200 OK", func(t *testing.T) {
		// Call woopra domain endpoint.
		assertNotSecuredGetResponseEquals("woopra-domain", "test woopra domain")
	})

	t.Run("get segment write key 200 OK", func(t *testing.T) {
		// Call segment write key endpoint.
		assertNotSecuredGetResponseEquals("segment-write-key", "test segment write key")
	})
}

func TestAuthConfig(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	t.Run("get authconfig 200 OK", func(t *testing.T) {
		// when & then
		// Call authconfig endpoint.
		wait.NewHTTPRequest().Method("GET").
			URL(route + "/api/v1/authconfig").
			RequireStatusCode(http.StatusOK).
			Execute(t)
	})
}

func TestSignupFails(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	t.Run("post signup error no token 401 Unauthorized", func(t *testing.T) {
		// when
		// Call signup endpoint without a token.
		mp, _ := wait.NewHTTPRequest().
			Method("POST").
			URL(route + "/api/v1/signup").
			RequireStatusCode(http.StatusUnauthorized).
			Execute(t)

		// then
		// Check token error.
		tokenErr := mp["error"]
		require.Equal(t, "no token found", tokenErr.(string))
	})
	t.Run("post signup error invalid token 401 Unauthorized", func(t *testing.T) {
		// when
		// Call signup endpoint with an invalid token.
		mp, _ := wait.NewHTTPRequest().
			Method("POST").
			URL(route + "/api/v1/signup").
			Token("1223123123").
			RequireStatusCode(http.StatusUnauthorized).
			Execute(t)

		// then
		// Check token error.
		tokenErr := mp["error"]
		require.Equal(t, "token contains an invalid number of segments", tokenErr.(string))
	})
	t.Run("post signup exp token 401 Unauthorized", func(t *testing.T) {
		// when
		emailAddress := uuid.Must(uuid.NewV4()).String() + "@acme.com"
		// Not identical to the token used in POST signup - should return resource not found.
		_, token1, err := authsupport.NewToken(
			authsupport.WithEmail(emailAddress),
			authsupport.WithExp(time.Now().Add(-60*time.Second)))
		require.NoError(t, err)
		mp, _ := wait.NewHTTPRequest().
			Method("POST").
			URL(route + "/api/v1/signup").
			Token(token1).
			RequireStatusCode(http.StatusUnauthorized).
			Execute(t)

		// then
		// Check token error.
		tokenErr := mp["error"]
		require.Contains(t, tokenErr.(string), "token is expired by ")
	})
	t.Run("get signup error no token 401 Unauthorized", func(t *testing.T) {
		// when
		// Call signup endpoint without a token.
		mp, _ := wait.NewHTTPRequest().
			Method("GET").
			URL(route + "/api/v1/signup").
			RequireStatusCode(http.StatusUnauthorized).
			Execute(t)

		// then
		// Check token error.
		tokenErr := mp["error"]
		require.Equal(t, "no token found", tokenErr.(string))
	})
	t.Run("get signup error invalid token 401 Unauthorized", func(t *testing.T) {
		// when
		// Call signup endpoint with an invalid token.
		mp, _ := wait.NewHTTPRequest().Method("GET").
			URL(route + "/api/v1/signup").
			Token("1223123123").
			RequireStatusCode(http.StatusUnauthorized).
			Execute(t)

		// then
		// Check token error.
		tokenErr := mp["error"]
		require.Equal(t, "token contains an invalid number of segments", tokenErr.(string))
	})
	t.Run("get signup exp token 401 Unauthorized", func(t *testing.T) {
		// when
		emailAddress := uuid.Must(uuid.NewV4()).String() + "@acme.com"
		// Not identical to the token used in POST signup - should return resource not found.
		_, token1, err := authsupport.NewToken(
			authsupport.WithEmail(emailAddress),
			authsupport.WithExp(time.Now().Add(-60*time.Second)))
		require.NoError(t, err)
		mp, _ := wait.NewHTTPRequest().
			Method("GET").
			URL(route + "/api/v1/signup").
			Token(token1).
			RequireStatusCode(http.StatusUnauthorized).
			Execute(t)

		// then
		// Check token error.
		tokenErr := mp["error"]
		require.Contains(t, tokenErr.(string), "token is expired by ")
	})
	t.Run("get signup 404 NotFound", func(t *testing.T) {
		// when
		// Get valid generated token for e2e tests. IAT claim is overridden
		// to avoid token used before issued error.
		// Not identical to the token used in POST signup - should return resource not found.
		_, token1, err := authsupport.NewToken(
			authsupport.WithEmail(uuid.Must(uuid.NewV4()).String()+"@acme.com"),
			authsupport.WithIAT(time.Now().Add(-60*time.Second)))
		require.NoError(t, err)

		// then
		// Call signup endpoint with a valid token.
		assertGetSignupReturnsNotFound(t, await, token1)
	})

	t.Run("get signup for crtadmin fails", func(t *testing.T) {
		// when
		// Get valid generated token for e2e tests. IAT claim is overridden
		// to avoid token used before issued error. Username claim is also
		// overridden to trigger error and ensure that usersignup is not created.
		emailAddress := uuid.Must(uuid.NewV4()).String() + "@acme.com"
		identity, token, err := authsupport.NewToken(
			authsupport.WithEmail(emailAddress),
			authsupport.WithPreferredUsername("test-crtadmin"))
		require.NoError(t, err)

		// Call signup endpoint with a valid token to initiate a signup process
		response, _ := wait.NewHTTPRequest().
			Method("POST").
			URL(route + "/api/v1/signup").
			Token(token).
			RequireStatusCode(http.StatusForbidden).
			Execute(t)
		require.Equal(t, "forbidden: failed to create usersignup for test-crtadmin", response["message"])
		require.Equal(t, "error creating UserSignup resource", response["details"])
		require.Equal(t, float64(403), response["code"])

		// then
		hostAwait := await.Host()
		hostAwait.WithRetryOptions(wait.TimeoutOption(time.Second * 15)).WaitAndVerifyThatUserSignupIsNotCreated(identity.ID.String())
	})
}

func TestSignupOK(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	hostAwait := await.Host()
	memberAwait := await.Member1()
	signupUser := func(token, email, userSignupName string, identity *commonauth.Identity) *toolchainv1alpha1.UserSignup {
		// Call signup endpoint with a valid token to initiate a signup process
		wait.NewHTTPRequest().
			Method("POST").
			URL(route + "/api/v1/signup").
			Token(token).
			RequireStatusCode(http.StatusAccepted).
			Execute(t)

		// Wait for the UserSignup to be created
		userSignup, err := hostAwait.WaitForUserSignup(userSignupName,
			wait.UntilUserSignupHasConditions(ConditionSet(Default(), PendingApproval())...),
			wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValuePending))
		require.NoError(t, err)
		cleanup.AddCleanTasks(hostAwait, userSignup)
		emailAnnotation := userSignup.Annotations[toolchainv1alpha1.UserSignupUserEmailAnnotationKey]
		assert.Equal(t, email, emailAnnotation)

		// Call get signup endpoint with a valid token and make sure it's pending approval
		assertGetSignupStatusPendingApproval(t, await, identity.Username, token)

		// Attempt to create same usersignup by calling post signup with same token should return an error
		mp, _ := wait.NewHTTPRequest().
			Method("POST").
			URL(route + "/api/v1/signup").
			Token(token).
			RequireStatusCode(http.StatusConflict).
			Execute(t)
		assert.Equal(t, fmt.Sprintf("Operation cannot be fulfilled on  \"\": UserSignup [id: %s; username: %s]. Unable to create UserSignup because there is already an active UserSignup with such ID",
			identity.ID, identity.Username), mp["message"])
		assert.Equal(t, "error creating UserSignup resource", mp["details"])

		userSignup, err = hostAwait.UpdateUserSignup(userSignup.Name, func(instance *toolchainv1alpha1.UserSignup) {
			// Approve usersignup.
			states.SetApprovedManually(instance, true)
			instance.Spec.TargetCluster = memberAwait.ClusterName
		})
		require.NoError(t, err)

		// Wait for the resources to be provisioned
		VerifyResourcesProvisionedForSignup(t, await, userSignup, "deactivate30", "base")

		// Call signup endpoint with same valid token to check if status changed to Provisioned now
		assertGetSignupStatusProvisioned(t, await, identity.Username, token)

		return userSignup
	}

	t.Run("test activation-deactivation workflow", func(t *testing.T) {
		// Get valid generated token for e2e tests. IAT claim is overridden
		// to avoid token used before issued error.
		emailAddress := uuid.Must(uuid.NewV4()).String() + "@acme.com"
		identity, token, err := authsupport.NewToken(authsupport.WithEmail(emailAddress))
		require.NoError(t, err)

		// Signup a new user
		userSignup := signupUser(token, emailAddress, identity.Username, identity)

		// Deactivate the usersignup
		userSignup, err = hostAwait.UpdateUserSignup(userSignup.Name, func(us *toolchainv1alpha1.UserSignup) {
			states.SetDeactivated(us, true)
		})
		require.NoError(t, err)
		_, err = hostAwait.WaitForUserSignup(userSignup.Name,
			wait.UntilUserSignupHasConditions(ConditionSet(Default(), ApprovedByAdmin(), DeactivatedWithoutPreDeactivation())...),
			wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValueDeactivated))
		require.NoError(t, err)

		// Now check that the reg-service treats the deactivated usersignup as nonexistent and returns 404
		assertGetSignupReturnsNotFound(t, await, token)

		// Re-activate the usersignup by calling the signup endpoint with the same token/user again
		signupUser(token, emailAddress, identity.Username, identity)
	})
}
func TestUserSignupFoundWhenNamedWithEncodedUsername(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	hostAwait := await.Host()

	// Create a token and identity to sign up with, but override the username with "arnold" so that we create a UserSignup
	// with that name
	emailAddress := "arnold@acme.com"
	_, token0, err := authsupport.NewToken(
		authsupport.WithEmail(emailAddress),
		// authsupport.WithSub(identity0.ID.String()),
		authsupport.WithPreferredUsername("arnold"))
	require.NoError(t, err)

	// when
	// Call the signup endpoint
	wait.NewHTTPRequest().
		Method("POST").
		URL(route + "/api/v1/signup").
		Token(token0).
		RequireStatusCode(http.StatusAccepted).
		Execute(t)

	// Wait for the UserSignup to be created
	userSignup, err := hostAwait.WaitForUserSignup("arnold",
		wait.UntilUserSignupHasConditions(ConditionSet(Default(), PendingApproval())...),
		wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValuePending))
	require.NoError(t, err)
	cleanup.AddCleanTasks(hostAwait, userSignup)
	emailAnnotation := userSignup.Annotations[toolchainv1alpha1.UserSignupUserEmailAnnotationKey]
	assert.Equal(t, emailAddress, emailAnnotation)

	// then
	// Call get signup endpoint with a valid token, however we will now override the claims to introduce the original
	// sub claim and set username as a separate claim, then we will make sure the UserSignup is returned correctly
	_, token0, err = authsupport.NewToken(
		authsupport.WithEmail(emailAddress),
		authsupport.WithPreferredUsername("arnold"))
	require.NoError(t, err)

	mp, mpStatus := wait.NewHTTPRequest().
		Method("GET").
		URL(route + "/api/v1/signup").
		Token(token0).
		RequireStatusCode(http.StatusOK).
		ParseResponse().
		Execute(t)
	assert.Equal(t, "", mp["compliantUsername"])
	assert.Equal(t, "arnold", mp["username"])
	require.IsType(t, false, mpStatus["ready"])
	assert.False(t, mpStatus["ready"].(bool))
	assert.Equal(t, "PendingApproval", mpStatus["reason"])
}

func TestPhoneVerification(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	route := await.Host().RegistrationServiceURL

	hostAwait := await.Host()
	// Create a token and identity to sign up with
	emailAddress := uuid.Must(uuid.NewV4()).String() + "@some.domain"
	identity0, token0, err := authsupport.NewToken(authsupport.WithEmail(emailAddress))
	require.NoError(t, err)

	// Call the signup endpoint
	wait.NewHTTPRequest().
		Method("POST").
		URL(route + "/api/v1/signup").
		Token(token0).
		RequireStatusCode(http.StatusAccepted).
		Execute(t)

	// Wait for the UserSignup to be created
	userSignup, err := hostAwait.WaitForUserSignup(identity0.Username,
		wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...),
		wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValueNotReady))
	require.NoError(t, err)
	cleanup.AddCleanTasks(hostAwait, userSignup)
	emailAnnotation := userSignup.Annotations[toolchainv1alpha1.UserSignupUserEmailAnnotationKey]
	assert.Equal(t, emailAddress, emailAnnotation)

	// Call get signup endpoint with a valid token and make sure verificationRequired is true
	mp, mpStatus := wait.NewHTTPRequest().
		Method("GET").
		URL(route + "/api/v1/signup").
		Token(token0).
		RequireStatusCode(http.StatusOK).
		Execute(t)
	assert.Equal(t, "", mp["compliantUsername"])
	assert.Equal(t, identity0.Username, mp["username"])
	require.IsType(t, false, mpStatus["ready"])
	assert.False(t, mpStatus["ready"].(bool))
	assert.Equal(t, "PendingApproval", mpStatus["reason"])
	require.True(t, mpStatus["verificationRequired"].(bool))

	// Confirm the status of the UserSignup is correct
	_, err = hostAwait.WaitForUserSignup(identity0.Username,
		wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...),
		wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValueNotReady))
	require.NoError(t, err)

	// Confirm that a MUR hasn't been created
	obj := &toolchainv1alpha1.MasterUserRecord{}
	err = hostAwait.Client.Get(context.TODO(), types.NamespacedName{Namespace: hostAwait.Namespace, Name: identity0.Username}, obj)
	require.Error(t, err)
	require.True(t, errors.IsNotFound(err))

	// Initiate the verification process
	wait.NewHTTPRequest().
		Method("PUT").
		URL(route + "/api/v1/signup/verification").
		Token(token0).
		Body(`{ "country_code":"+61", "phone_number":"408999999" }`).
		RequireStatusCode(http.StatusNoContent).
		Execute(t)

	// Retrieve the updated UserSignup
	userSignup, err = hostAwait.WaitForUserSignup(identity0.Username)
	require.NoError(t, err)

	// Confirm there is a verification code annotation value, and store it in a variable
	verificationCode := userSignup.Annotations[toolchainv1alpha1.UserSignupVerificationCodeAnnotationKey]
	require.NotEmpty(t, verificationCode)

	// Confirm the expiry time has been set
	require.NotEmpty(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationExpiryAnnotationKey])

	// Attempt to verify with an incorrect verification code
	wait.NewHTTPRequest().
		Method("GET").
		URL(route + "/api/v1/signup/verification/invalid").
		Token(token0).
		RequireStatusCode(http.StatusForbidden).
		Execute(t)

	// Retrieve the updated UserSignup
	userSignup, err = hostAwait.WaitForUserSignup(identity0.Username)
	require.NoError(t, err)

	// Check attempts has been incremented
	require.NotEmpty(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationAttemptsAnnotationKey])

	// Confirm the verification code has not changed
	require.Equal(t, verificationCode, userSignup.Annotations[toolchainv1alpha1.UserSignupVerificationCodeAnnotationKey])

	// Verify with the correct code
	wait.NewHTTPRequest().
		Method("GET").
		URL(route + fmt.Sprintf("/api/v1/signup/verification/%s",
			userSignup.Annotations[toolchainv1alpha1.UserSignupVerificationCodeAnnotationKey])).
		Token(token0).
		RequireStatusCode(http.StatusOK).
		Execute(t)

	// Retrieve the updated UserSignup
	userSignup, err = hostAwait.WaitForUserSignup(identity0.Username,
		wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValuePending))
	require.NoError(t, err)

	// Confirm all unrequired verification-related annotations have been removed
	require.Empty(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationExpiryAnnotationKey])
	require.Empty(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationAttemptsAnnotationKey])
	require.Empty(t, userSignup.Annotations[toolchainv1alpha1.UserSignupVerificationCodeAnnotationKey])
	require.Empty(t, userSignup.Annotations[toolchainv1alpha1.UserSignupVerificationTimestampAnnotationKey])
	require.Empty(t, userSignup.Annotations[toolchainv1alpha1.UserSignupVerificationCounterAnnotationKey])
	require.Empty(t, userSignup.Annotations[toolchainv1alpha1.UserSignupVerificationInitTimestampAnnotationKey])

	// Call get signup endpoint with a valid token and make sure it's pending approval
	mp, mpStatus = wait.NewHTTPRequest().
		Method("GET").
		URL(route + "/api/v1/signup").
		Token(token0).
		RequireStatusCode(http.StatusOK).
		ParseResponse().
		Execute(t)
	assert.Equal(t, "", mp["compliantUsername"])
	assert.Equal(t, identity0.Username, mp["username"])
	require.IsType(t, false, mpStatus["ready"])
	assert.False(t, mpStatus["ready"].(bool))
	assert.Equal(t, "PendingApproval", mpStatus["reason"])
	require.False(t, mpStatus["verificationRequired"].(bool))

	userSignup, err = hostAwait.UpdateUserSignup(userSignup.Name, func(instance *toolchainv1alpha1.UserSignup) {
		// Now approve the usersignup.
		states.SetApprovedManually(instance, true)
	})
	require.NoError(t, err)

	// Confirm the MasterUserRecord is provisioned
	_, err = hostAwait.WaitForMasterUserRecord(identity0.Username, wait.UntilMasterUserRecordHasCondition(Provisioned()))
	require.NoError(t, err)

	// Retrieve the UserSignup from the GET endpoint
	_, mpStatus = wait.NewHTTPRequest().
		Method("GET").
		URL(route + "/api/v1/signup").
		Token(token0).
		RequireStatusCode(http.StatusOK).
		ParseResponse().
		Execute(t)

	// Confirm that VerificationRequired is no longer true
	require.False(t, mpStatus["verificationRequired"].(bool))

	// Create another token and identity to sign up with
	otherEmailValue := uuid.Must(uuid.NewV4()).String() + "@other.domain"
	otherIdentity, otherToken, err := authsupport.NewToken(authsupport.WithEmail(otherEmailValue))
	require.NoError(t, err)

	// Call the signup endpoint
	wait.NewHTTPRequest().
		Method("POST").
		URL(route + "/api/v1/signup").
		Token(otherToken).
		RequireStatusCode(http.StatusAccepted).
		Execute(t)

	// Wait for the UserSignup to be created
	otherUserSignup, err := hostAwait.WaitForUserSignup(otherIdentity.Username,
		wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...),
		wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValueNotReady))
	require.NoError(t, err)
	cleanup.AddCleanTasks(hostAwait, otherUserSignup)
	otherEmailAnnotation := otherUserSignup.Annotations[toolchainv1alpha1.UserSignupUserEmailAnnotationKey]
	assert.Equal(t, otherEmailValue, otherEmailAnnotation)

	// Initiate the verification process using the same phone number as previously
	responseMap, _ := wait.NewHTTPRequest().
		Method("PUT").
		URL(route + "/api/v1/signup/verification").
		Token(otherToken).
		Body(`{ "country_code":"+61", "phone_number":"408999999" }`).
		RequireStatusCode(http.StatusForbidden).
		Execute(t)

	require.NotEmpty(t, responseMap)
	require.Equal(t, float64(http.StatusForbidden), responseMap["code"], "code not found in response body map %s", responseMap)

	require.Equal(t, "Forbidden", responseMap["status"])
	require.Equal(t, "phone number already in use: cannot register using phone number: +61408999999", responseMap["message"])
	require.Equal(t, "phone number already in use", responseMap["details"])

	// Retrieve the updated UserSignup
	otherUserSignup, err = hostAwait.WaitForUserSignup(otherIdentity.Username)
	require.NoError(t, err)

	// Confirm there is no verification code annotation value
	require.Empty(t, otherUserSignup.Annotations[toolchainv1alpha1.UserSignupVerificationCodeAnnotationKey])

	// Retrieve the current UserSignup
	userSignup, err = hostAwait.WaitForUserSignup(userSignup.Name)
	require.NoError(t, err)

	userSignup, err = hostAwait.UpdateUserSignup(userSignup.Name, func(instance *toolchainv1alpha1.UserSignup) {
		// Now mark the original UserSignup as deactivated
		states.SetDeactivated(instance, true)
	})
	require.NoError(t, err)

	// Ensure the UserSignup is deactivated
	_, err = hostAwait.WaitForUserSignup(userSignup.Name,
		wait.UntilUserSignupHasConditions(ConditionSet(Default(), ApprovedByAdmin(), ManuallyDeactivated())...))
	require.NoError(t, err)

	// Now attempt the verification again
	wait.NewHTTPRequest().
		Method("PUT").
		URL(route + "/api/v1/signup/verification").
		Token(otherToken).
		Body(`{ "country_code":"+61", "phone_number":"408999999" }`).
		RequireStatusCode(http.StatusNoContent).
		Execute(t)

	// Retrieve the updated UserSignup again
	otherUserSignup, err = hostAwait.WaitForUserSignup(otherIdentity.Username)
	require.NoError(t, err)

	// Confirm there is now a verification code annotation value
	require.NotEmpty(t, otherUserSignup.Annotations[toolchainv1alpha1.UserSignupVerificationCodeAnnotationKey])
}

func TestActivationCodeVerification(t *testing.T) {
	// given
	t.Parallel()
	await := WaitForDeployments(t)
	hostAwait := await.Host()
	route := hostAwait.RegistrationServiceURL

	t.Run("verification successful", func(t *testing.T) {
		// given
		event := testsocialevent.NewSocialEvent(hostAwait.Namespace, commonsocialevent.NewName(),
			testsocialevent.WithUserTier("deactivate80"),
			testsocialevent.WithSpaceTier("base1ns6didler"))
		err := hostAwait.CreateWithCleanup(context.TODO(), event)
		require.NoError(t, err)
		userSignup, token := signup(t, hostAwait)

		// when call verification endpoint with a valid activation code
		wait.NewHTTPRequest().
			Method("POST").
			URL(route + "/api/v1/signup/verification/activation-code").
			Token(token).
			Body(fmt.Sprintf(`{"code":"%s"}`, event.Name)).
			RequireStatusCode(http.StatusOK).
			Execute(t)

		// then
		// ensure the UserSignup is in "pending approval" condition,
		// because in these series of parallel tests, automatic approval is disabled ¯\_(ツ)_/¯
		_, err = hostAwait.WaitForUserSignup(userSignup.Name,
			wait.UntilUserSignupHasLabel(toolchainv1alpha1.SocialEventUserSignupLabelKey, event.Name),
			wait.UntilUserSignupHasConditions(ConditionSet(Default(), PendingApproval())...))
		require.NoError(t, err)
		// explicitly approve the usersignup (see above, config for parallel test has automatic approval disabled)
		userSignup, err = hostAwait.UpdateUserSignup(userSignup.Name, func(us *toolchainv1alpha1.UserSignup) {
			states.SetApprovedManually(us, true)
		})
		require.NoError(t, err)
		t.Logf("user signup '%s' approved", userSignup.Name)

		// check that the MUR and Space are configured as expected
		// Wait for the UserSignup to have the desired state
		userSignup, err = hostAwait.WaitForUserSignup(userSignup.Name,
			wait.UntilUserSignupHasCompliantUsername())
		require.NoError(t, err)
		mur, err := hostAwait.WaitForMasterUserRecord(userSignup.Status.CompliantUsername,
			wait.UntilMasterUserRecordHasTierName(event.Spec.UserTier),
			wait.UntilMasterUserRecordHasCondition(Provisioned()))
		require.NoError(t, err)
		assert.Equal(t, event.Spec.UserTier, mur.Spec.TierName)
		_, err = hostAwait.WaitForSpace(userSignup.Status.CompliantUsername,
			wait.UntilSpaceHasTier(event.Spec.SpaceTier),
			wait.UntilSpaceHasConditions(Provisioned()),
		)
		require.NoError(t, err)

		// also check that the SocialEvent status was updated accordingly
		_, err = hostAwait.WaitForSocialEvent(event.Name, wait.UntilSocialEventHasActivationCount(1))
		require.NoError(t, err)
	})

	t.Run("verification failed", func(t *testing.T) {

		t.Run("unknown code", func(t *testing.T) {
			// given
			userSignup, token := signup(t, hostAwait)

			// when call verification endpoint with a valid activation code
			wait.NewHTTPRequest().
				Method("POST").
				URL(route + "/api/v1/signup/verification/activation-code").
				Token(token).
				Body(fmt.Sprintf(`{"code":"%s"}`, "unknown")).
				RequireStatusCode(http.StatusForbidden).
				Execute(t)

			// then
			// ensure the UserSignup is not approved yet
			userSignup, err := hostAwait.WaitForUserSignup(userSignup.Name,
				wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...))
			require.NoError(t, err)
			assert.Equal(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationAttemptsAnnotationKey], "1")
		})

		t.Run("over capacity", func(t *testing.T) {
			// given
			event := testsocialevent.NewSocialEvent(hostAwait.Namespace, commonsocialevent.NewName(),
				testsocialevent.WithUserTier("deactivate80"),
				testsocialevent.WithSpaceTier("base1ns6didler"))
			err := hostAwait.CreateWithCleanup(context.TODO(), event)
			require.NoError(t, err)
			event, err = hostAwait.WaitForSocialEvent(event.Name) // need to reload event
			require.NoError(t, err)
			event.Status.ActivationCount = event.Spec.MaxAttendees // activation count identical to `MaxAttendees`
			err = hostAwait.Client.Status().Update(context.TODO(), event)
			require.NoError(t, err)

			userSignup, token := signup(t, hostAwait)

			// when call verification endpoint with a valid activation code
			wait.NewHTTPRequest().
				Method("POST").
				URL(route + "/api/v1/signup/verification/activation-code").
				Token(token).
				Body(fmt.Sprintf(`{"code":"%s"}`, event.Name)).
				RequireStatusCode(http.StatusForbidden).
				Execute(t)

			// then
			// ensure the UserSignup is not approved yet
			userSignup, err = hostAwait.WaitForUserSignup(userSignup.Name,
				wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...))
			require.NoError(t, err)
			assert.Equal(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationAttemptsAnnotationKey], "1")
		})

		t.Run("not opened yet", func(t *testing.T) {
			// given
			event := testsocialevent.NewSocialEvent(hostAwait.Namespace, commonsocialevent.NewName(), testsocialevent.WithStartTime(time.Now().Add(time.Hour))) // not open yet
			err := hostAwait.CreateWithCleanup(context.TODO(), event)
			require.NoError(t, err)
			userSignup, token := signup(t, hostAwait)

			// when call verification endpoint with a valid activation code
			wait.NewHTTPRequest().
				Method("POST").
				URL(route + "/api/v1/signup/verification/activation-code").
				Token(token).
				Body(fmt.Sprintf(`{"code":"%s"}`, event.Name)).
				RequireStatusCode(http.StatusForbidden).
				Execute(t)

			// then
			// ensure the UserSignup is not approved yet
			userSignup, err = hostAwait.WaitForUserSignup(userSignup.Name,
				wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...))
			require.NoError(t, err)
			assert.Equal(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationAttemptsAnnotationKey], "1")
		})

		t.Run("already closed", func(t *testing.T) {
			// given
			event := testsocialevent.NewSocialEvent(hostAwait.Namespace, commonsocialevent.NewName(), testsocialevent.WithEndTime(time.Now().Add(-time.Hour))) // already closd
			err := hostAwait.CreateWithCleanup(context.TODO(), event)
			require.NoError(t, err)
			userSignup, token := signup(t, hostAwait)

			// when call verification endpoint with a valid activation code
			wait.NewHTTPRequest().
				Method("POST").
				URL(route + "/api/v1/signup/verification/activation-code").
				Token(token).
				Body(fmt.Sprintf(`{"code":"%s"}`, event.Name)).
				RequireStatusCode(http.StatusForbidden).
				Execute(t)

			// then
			// ensure the UserSignup is approved
			userSignup, err = hostAwait.WaitForUserSignup(userSignup.Name,
				wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...))
			require.NoError(t, err)
			assert.Equal(t, userSignup.Annotations[toolchainv1alpha1.UserVerificationAttemptsAnnotationKey], "1")
		})
	})
}

func signup(t *testing.T, hostAwait *wait.HostAwaitility) (*toolchainv1alpha1.UserSignup, string) {
	route := hostAwait.RegistrationServiceURL

	// Create a token and identity to sign up with
	identity := commonauth.NewIdentity()
	emailValue := identity.Username + "@some.domain"
	emailClaim := commonauth.WithEmailClaim(emailValue)
	token, err := commonauth.GenerateSignedE2ETestToken(*identity, emailClaim)
	require.NoError(t, err)

	// Call the signup endpoint
	wait.NewHTTPRequest().
		Method("POST").
		URL(route + "/api/v1/signup").
		Token(token).
		RequireStatusCode(http.StatusAccepted).
		Execute(t)

	// Wait for the UserSignup to be created
	userSignup, err := hostAwait.WaitForUserSignup(identity.Username,
		wait.UntilUserSignupHasConditions(ConditionSet(Default(), VerificationRequired())...),
		wait.UntilUserSignupHasStateLabel(toolchainv1alpha1.UserSignupStateLabelValueNotReady))
	require.NoError(t, err)
	cleanup.AddCleanTasks(hostAwait, userSignup)
	emailAnnotation := userSignup.Annotations[toolchainv1alpha1.UserSignupUserEmailAnnotationKey]
	assert.Equal(t, emailValue, emailAnnotation)
	return userSignup, token
}

func assertGetSignupStatusProvisioned(t *testing.T, await wait.Awaitilities, username, bearerToken string) {
	hostAwait := await.Host()
	memberAwait := await.Member1()
	mp := hostAwait.WaitForUserSignupReadyInRegistrationService(t, username, bearerToken)
	assert.Equal(t, username, mp["compliantUsername"])
	assert.Equal(t, username, mp["username"])
	assert.Equal(t, memberAwait.GetConsoleURL(), mp["consoleURL"])
	memberCluster, found, err := hostAwait.GetToolchainCluster(cluster.Member, memberAwait.Namespace, nil)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, memberCluster.Spec.APIEndpoint, mp["apiEndpoint"])
	assert.Equal(t, hostAwait.APIProxyURL, mp["proxyURL"])
}

func assertGetSignupStatusPendingApproval(t *testing.T, await wait.Awaitilities, username, bearerToken string) {
	route := await.Host().RegistrationServiceURL
	mp, mpStatus := wait.NewHTTPRequest().
		Method("GET").
		URL(route + "/api/v1/signup").
		Token(bearerToken).
		RequireStatusCode(http.StatusOK).
		ParseResponse().
		Execute(t)
	assert.Equal(t, username, mp["username"])
	require.IsType(t, false, mpStatus["ready"])
	assert.False(t, mpStatus["ready"].(bool))
	assert.Equal(t, "PendingApproval", mpStatus["reason"])
}

func assertGetSignupReturnsNotFound(t *testing.T, await wait.Awaitilities, bearerToken string) {
	route := await.Host().RegistrationServiceURL
	wait.NewHTTPRequest().
		Method("GET").
		URL(route + "/api/v1/signup").
		Token(bearerToken).
		RequireStatusCode(http.StatusNotFound).Execute(t)
}
