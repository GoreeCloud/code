// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"forgejo.org/models/db"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/util"

	gouuid "github.com/google/uuid"
	"xorm.io/builder"
)

// An Authorized Integration allow users to define external systems which can generate JSON Web Tokens (JWTs) that
// Forgejo will trust in order to perform API access on behalf of a user defined by the UserID field.
//
// When a JWT is received by Forgejo, the issuer (iss) and audience (aud) claims are used to lookup an authorized
// integration with an exact match.  Together these fields serve as a unique key for the authorized issuer.  Duplicates
// cannot be permitted because we would not know which user to authenticate the JWT as.
type AuthorizedIntegration struct {
	ID int64 `xorm:"pk autoincr"`

	UserID           int64            `xorm:"NOT NULL REFERENCES(user, id)"`
	Scope            AccessTokenScope `xorm:"NOT NULL"`
	ResourceAllRepos bool             `xorm:"NOT NULL"` // flag for whether AuthorizedIntegrationResourceRepo instances will limit the resources this access token can access (false) or won't limit them (true).

	Name        string // short name for lists of authorized integrations
	Description string `xorm:"LONGTEXT"` // long description, optional to document relevant details of the integration

	// Which UI to use for view/edit of this Authorized Integration.  Authorized Integrations' functional behaviour is
	// defined by other fields, such as the Issuer, Audience, ClaimRules.  The UI field only defines how this record can
	// be interacted with by the user in order to provide user-friendly access for specific systems -- like Forgejo
	// Actions.  Within the potential scope of the UI is any user interaction with the Authorized Integration -- web
	// create, read, update, API, CLI.
	//
	// The UI field must never be used to make functional decisions about evaluating JWTs.  It must always be possible
	// to convert an Authorized Integration to a "generic" UI (for customization that the UI doesn't support).  The
	// intent of this design is that, the Authorized Integration system is complicated in its claim rules, but they
	// always fully define the behaviour in a transparent manner.
	UI AuthorizedIntegrationUI `xorm:"NOT NULL default('generic')"`

	// Exact-match `iss` claim of the JWT
	Issuer string `xorm:"NOT NULL UNIQUE(s)"`
	// Exact-match `aud` claim of the JWT
	Audience   string      `xorm:"NOT NULL UNIQUE(s)"`
	ClaimRules *ClaimRules `xorm:"NOT NULL JSON"`

	CreatedUnix timeutil.TimeStamp `xorm:"NOT NULL created"`
	UpdatedUnix timeutil.TimeStamp `xorm:"NOT NULL updated"`
}

func init() {
	db.RegisterModel(new(AuthorizedIntegration))
}

// An [AuthorizedIntegration] can validate the claims in a JWT against a set of rules defined by this structure.
//
// JWTs can contain any number of claims, which are represented as a JSON object.  A small number of common claims are
// described in RFC7519 (sec 4.1) which defines JWTs, but most claims are entirely arbitrarily defined by the JWT
// issuer.
//
// For example, eg. a claim may be {"sub": "repo:coolguy/forgejo-runner-testrepo:pull_request"} indicating that an OIDC
// token was received from an Actions execution in a specific repo on a specific event.
//
// Validating the claims from a JWT issuer is a critical part of creating a secure [AuthorizedIssuer].  For example,
// assume that we receive a JWT from a public hosting platform like Codeberg.  We will validate that it is a claim
// created by the correct Issuer, Codeberg -- but anyone can do that through Forgejo Actions.  We will validate that it
// has the correct audience -- but that's an *input* to Forgejo Actions, so anyone can create a claim on Codeberg with
// an arbitrary audience.  The rest of the claims contain the critical information about who ran a Forgejo Action, on
// which repository, and in response to which events, and those must be validated to ensure that an authorized issuer is
// correctly authorized.
//
// Following that an example, a minimum claim rule that would be required for securely using Forgejo Actions would be
// something like:
//
//	{
//	  "rules": [{
//	     "claim": "sub",
//	     "compare": "eq",
//	     "value": "repo:forgejo/website:pull_request"
//	  }]
//	}
//
// This defines a single rule which says that the `sub` claim must be exactly equal to
// "repo:forgejo/website:pull_request".  Forgejo Actions would generate this subject when an Action is running on the
// repo forgejo/website in response to the pull_request event.
//
// Some JWT claims are JSON objects.  The [ClaimNested] comparison operator can be used to define rules that inspect the
// object within a claim.  For example, AWS STS generates a claim "https://sts.amazonaws.com/": {...} with values inside
// an object, like "aws_account".  A nested claim can inspect those values:
//
//	{
//	  "rules":[{
//	    "claim": "https://sts.amazonaws.com/",
//	    "compare": "nest",
//	    "nested": {"rules":[
//	      {"claim": "aws_account", "compare": "eq", "value": "1234567890"},
//	      {"claim": "lambda_source_function_arn", "compare": "eq", "value": "arn:aws:lambda:ca-central-1:1234567890:function:forgejo-oidc-accepting-test"}
//	    ]}
//	  }
//
// ]}
//
// This defines a rule that looks into the "https://sts..." claim and verifies the "aws_account" and
// "lambda_source_function_arn" keys match specific known values.
type ClaimRules struct {
	Rules []ClaimRule `json:"rules"`
}

// Defines a single rule that will check the value of one JWT claim.
type ClaimRule struct {
	// The target claim, eg. "sub"
	Claim string `json:"claim"`
	// Comparison rule to use on this claim
	Comparison ClaimComparison `json:"compare"`

	// For Comparison of ClaimEqual or ClaimGlob, the specific value or glob to match against
	Value string `json:"value,omitempty"`

	// For Comparison of ClaimIn or ClaimGlobIn, an array of values to match against
	Values []string `json:"values,omitempty"`

	// For ClaimNested, the rules to apply to the nested object
	Nested *ClaimRules `json:"nested,omitempty"`
}

type ClaimComparison string

const (
	ClaimEqual  ClaimComparison = "eq"      // exactly equal claim
	ClaimIn     ClaimComparison = "in"      // exactly equal any of the options in a list
	ClaimGlob   ClaimComparison = "glob"    // glob match complete claim string
	ClaimGlobIn ClaimComparison = "glob-in" // glob match any of the options in a list
	ClaimNested ClaimComparison = "nest"    // recurse into a claim that is an map[string]any with it's own data fields
)

type AuthorizedIntegrationUI string

const (
	// Generic UI which allows the user to view and edit claim rules directly to support integrations that Forgejo
	// doesn't have a user-friendly UI to support.
	AuthorizedIntegrationUIGeneric AuthorizedIntegrationUI = "generic"

	// UI specific to Actions that are running on this local Forgejo instance accessing itself.
	AuthorizedIntegrationUIForgejoActionsLocal AuthorizedIntegrationUI = "forgejo-actions-local"
)

func GetAuthorizedIntegration(ctx context.Context, issuer, audience string) (*AuthorizedIntegration, error) {
	var ai AuthorizedIntegration
	found, err := db.GetEngine(ctx).Where("issuer = ? AND audience = ?", issuer, audience).Get(&ai)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, util.ErrNotExist
	}
	return &ai, nil
}

func GetAuthorizedIntegrationByUI(ctx context.Context, ownerID int64, aiUI AuthorizedIntegrationUI, aiID int64) (*AuthorizedIntegration, error) {
	var ai AuthorizedIntegration
	found, err := db.GetEngine(ctx).Where("id = ? AND user_id = ? AND ui = ?", aiID, ownerID, aiUI).Get(&ai)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, util.ErrNotExist
	}
	return &ai, nil
}

func InsertAuthorizedIntegration(ctx context.Context, ai *AuthorizedIntegration) error {
	if ai.Audience != "" {
		return errors.New("audience cannot be provided, and must be generated by NewAuthorizedIntegration")
	} else if err := ai.generateAudience(); err != nil {
		return err
	}
	_, err := db.GetEngine(ctx).Insert(ai)
	return err
}

func UpdateAuthorizedIntegration(ctx context.Context, ai *AuthorizedIntegration) error {
	// NoAutoTime -- UpdatedUnix is used to track the last used time, don't update it when editing
	// AllCols -- ensure ResourceAllRepo can be set to false
	rowsImpacted, err := db.GetEngine(ctx).ID(ai.ID).NoAutoTime().AllCols().Update(ai)
	if rowsImpacted == 0 {
		return fmt.Errorf("authorized integration update affected 0 records: %w", util.ErrNotExist)
	} else if rowsImpacted != 1 {
		return fmt.Errorf("authorized integration update affected %d records", rowsImpacted)
	}
	return err
}

// Bump the UpdatedUnix field of this authorized integration to now, tracking when it was last used for authentication.
// To reduce database write workload, this is only tracked by one-minute intervals -- the UPDATE statement conditionally
// avoids writes.
func (ai *AuthorizedIntegration) UpdateLastUsed(ctx context.Context) error {
	newTime := timeutil.TimeStampNow()
	cnt, err := db.GetEngine(ctx).
		Table(&AuthorizedIntegration{}).
		Where(builder.Eq{"id": ai.ID}).
		Where(builder.Lt{"updated_unix": newTime.AddDuration(-1 * time.Minute)}).
		NoAutoTime().
		Update(map[string]any{"updated_unix": newTime})
	if cnt == 1 {
		ai.UpdatedUnix = newTime
	}
	return err
}

// Generates the `aud` claim that the remote JWT generator must use to match this authorized integration.  The `aud`
// claim is an arbitrary value in a JWT claim, but Forgejo is faced with a few hard and soft requirements:
//
//   - Hard requirement: each authorized integration must have a unique `aud`, as it is used to find the DB record that
//     authenticates a request.
//   - If authentication is failing, being able to inspect the `aud` claim can be useful to identify the intent.
//   - Inspection should have a stable meaning -- eg. if it included the username, and the user was renamed, the `aud`
//     value which can't be changed would continue to reference the old username causing confusion when inspecting it.
//   - Forgejo & GitHub Actions uses a URL $ACTIONS_ID_TOKEN_REQUEST_URL&audience=... to generate a JWT for the running
//     action, so it should only consist of safe characters for URL encoding.
//   - It should be relatively short, as it's encoded into the JWT and increases its size.
//
// Meeting these requirements decently well is a combination of the owner's ID, a guid, and a "u:" prefix that makes the
// fact that it's an `aud` claim value a little bit identifiable.
func (ai *AuthorizedIntegration) generateAudience() error {
	if ai.UserID == 0 {
		return errors.New("UserID must be initialized")
	}
	ai.Audience = fmt.Sprintf("u:%d:%s", ai.UserID, gouuid.New().String())
	return nil
}

func (ai *AuthorizedIntegration) HasRecentActivity() bool {
	return ai.HasBeenUsed() && ai.UpdatedUnix.AddDuration(7*24*time.Hour) > timeutil.TimeStampNow()
}

func (ai *AuthorizedIntegration) HasBeenUsed() bool {
	return ai.UpdatedUnix > ai.CreatedUnix
}

type ListAuthorizedIntegrationOptions struct {
	db.ListOptions
	UserID optional.Option[int64]
}

func (opts ListAuthorizedIntegrationOptions) ToConds() builder.Cond {
	cond := builder.NewCond()
	if has, userID := opts.UserID.Get(); has {
		cond = cond.And(builder.Eq{"user_id": userID})
	}
	return cond
}

func (opts ListAuthorizedIntegrationOptions) ToOrders() string {
	return "created_unix DESC"
}

func ParseAuthorizedIntegrationUI(ui string) (AuthorizedIntegrationUI, error) {
	switch ui {
	case string(AuthorizedIntegrationUIGeneric):
		return AuthorizedIntegrationUIGeneric, nil
	case string(AuthorizedIntegrationUIForgejoActionsLocal):
		return AuthorizedIntegrationUIForgejoActionsLocal, nil
	}
	return AuthorizedIntegrationUI(""), fmt.Errorf("invalid authorized integration UI: %q", ui)
}

// Delete an authorized integration by ID.  Must only succeed if the authorized integration identified is owned by the
// user provided.
func DeleteAuthorizedIntegrationByID(ctx context.Context, id, userID int64) error {
	return db.WithTx(ctx, func(ctx context.Context) error {
		// Delete doesn't take into account userID, but will be rolled back by the transaction if the user ID isn't
		// correct.  Needs to occur first due to foreign key.
		if err := db.DeleteBeans(ctx,
			&AuthorizedIntegResourceRepo{IntegID: id},
		); err != nil {
			return fmt.Errorf("DeleteBeans: %w", err)
		}

		cnt, err := db.GetEngine(ctx).
			ID(id).
			Delete(&AuthorizedIntegration{
				UserID: userID,
			})
		if err != nil {
			return err
		} else if cnt != 1 {
			return fmt.Errorf("authorized integration %d does not exist: %w", id, util.ErrNotExist)
		}
		return nil
	})
}
