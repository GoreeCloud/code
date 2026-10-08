package container

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"forgejo.org/models/db"
	"forgejo.org/models/packages"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/require"
)

func TestGetCleanupTargets(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	ctx := db.DefaultContext

	createPackageCleanupRule := func(t *testing.T, keepCount, removeDays int) *packages.PackageCleanupRule {
		t.Helper()

		pcr := packages.PackageCleanupRule{
			Enabled:    true,
			OwnerID:    2001,
			Type:       packages.TypeContainer,
			KeepCount:  keepCount,
			RemoveDays: removeDays,
		}
		_, err := db.GetEngine(ctx).Insert(&pcr)
		require.NoError(t, err)
		return &pcr
	}

	createContainerVersions := func(t *testing.T, name string, count int) {
		t.Helper()

		p := packages.Package{
			OwnerID:   2001,
			Name:      name,
			LowerName: name,
			Type:      packages.TypeContainer,
		}
		_, err := db.GetEngine(ctx).Insert(&p)
		require.NoError(t, err)

		for i := range count {
			version := fmt.Sprintf("0.%d.0", i+1)
			created := time.Now().
				Add(-720 * time.Hour).
				Add(time.Duration(count-i) * time.Hour * -1).
				Unix()

			// Create the package version for the amd64 variant of a multi-platform OCI image
			platformAmd64Hash := sha256.New()
			platformAmd64Hash.Write([]byte(version + "amd64"))
			platformAmd64Version := "sha256:" + hex.EncodeToString(platformAmd64Hash.Sum(nil))
			platformAmd64PackageVersion := packages.PackageVersion{
				PackageID:    p.ID,
				Version:      platformAmd64Version,
				LowerVersion: platformAmd64Version,
				CreatedUnix:  timeutil.TimeStamp(created),
			}
			_, err = db.GetEngine(ctx).NoAutoTime().Insert(&platformAmd64PackageVersion)
			require.NoError(t, err)

			// Create the package version for the arm64 variant of a multi-platform OCI image
			platformArm64Hash := sha256.New()
			platformArm64Hash.Write([]byte(version + "arm64"))
			platformArm64Version := "sha256:" + hex.EncodeToString(platformArm64Hash.Sum(nil))
			platformArm64PackageVersion := packages.PackageVersion{
				PackageID:    p.ID,
				Version:      platformArm64Version,
				LowerVersion: platformArm64Version,
				CreatedUnix:  timeutil.TimeStamp(created),
			}
			_, err = db.GetEngine(ctx).NoAutoTime().Insert(&platformArm64PackageVersion)
			require.NoError(t, err)

			// Create the package version for tagged manifest of a multi-platform OCI image
			v := packages.PackageVersion{
				PackageID:    p.ID,
				Version:      version,
				LowerVersion: version,
				CreatedUnix:  timeutil.TimeStamp(created),
			}
			_, err = db.GetEngine(ctx).NoAutoTime().Insert(&v)
			require.NoError(t, err)
		}
	}

	t.Run("keeps the last five versions of multi-platform container images", func(t *testing.T) {
		pcr := createPackageCleanupRule(t, 5, 7)
		// Create versions 0.1.0 to 0.6.0
		createContainerVersions(t, "unit/test", 6)
		targets, err := GetCleanupTargets(ctx, pcr, true)
		require.NoError(t, err)
		require.Len(t, targets, 1)
		require.Equal(t, "0.1.0", targets[0].PackageVersion.LowerVersion)
	})
}
