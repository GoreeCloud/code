// Copyright 2026 The Forgejo Authors
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// web_src/css/org.css
// templates/org/member/members.tmpl
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';

test.use({user: 'user2'});

test('Toggle visibility', async ({page}) => {
  page.goto('/org/org3/members');

  // Button "Make hidden" for user2's row
  const hideUser2 = page.locator('.link-action[data-url="/org/org3/members/action/private?uid=2"]');
  // Button "Make visible" for user2's row
  const showUser2 = page.locator('.link-action[data-url="/org/org3/members/action/public?uid=2"]');

  await expect(hideUser2).toBeVisible();
  await expect(showUser2).toBeHidden();
  await hideUser2.click();

  // Button action was flipped
  await expect(hideUser2).toBeHidden();
  await expect(showUser2).toBeVisible();

  // Revert for repeatability
  await showUser2.click();
});

test('Leave org', async ({page}) => {
  page.goto('/org/org3/members');

  // Button "Leave" for user2's row
  const leaveButton = page.locator('.delete-button[data-url="/org/org3/members/action/leave"]');

  // Click the button
  await leaveButton.click();

  // A confirmation modal will appear
  await expect(page.locator('.modal#leave-organization')).toBeVisible();

  // Proceed leaving the org
  await page.locator('.modal#leave-organization .actions button.ok').click();

  // Getting error is enough to know that the correct request went though
  await expect(page.locator('.flash-error').getByText('You cannot remove the last user from the "owners" team.')).toBeVisible();
});
