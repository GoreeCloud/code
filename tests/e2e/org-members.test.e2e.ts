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
  await page.goto('/org/org3/members');

  // Button "Make hidden" for user2's row
  const hideUser2 = page.locator('.link-action[data-url="/org/org3/members/action/private?uid=2"]');
  // Button "Make visible" for user2's row
  const showUser2 = page.locator('.link-action[data-url="/org/org3/members/action/public?uid=2"]');

  if (await hideUser2.isHidden()) {
    // Bring page to a consistent state before testing
    await showUser2.click();
    await expect(hideUser2).toBeVisible();
    await expect(showUser2).toBeHidden();
  }

  // Click button to make user2 hidden
  await hideUser2.click();
  await page.waitForLoadState();

  // Verify that button action was flipped
  await expect(hideUser2).toBeHidden();
  await expect(showUser2).toBeVisible();
});

test('Leave org', async ({page}) => {
  await page.goto('/org/org3/members');

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

test('Add and remove a new member to the org', async ({page}) => {
  await page.goto('/org/org3/members');

  // Click the "Add member" button
  const newMemberButton = page.locator('#add-org-member-button');
  await newMemberButton.click();

  // A modal dialog appears
  await expect(page.locator('#add-member-modal')).toBeVisible();

  // Fill in the name of the user to add
  await page.locator('#search-user-box input').fill('user5');
  // Pick the auto-complete suggestion
  await page.locator('#search-user-box .results a.result').click();

  // Choose some teams
  await page.locator('#add-member-team_2').click();
  await page.locator('#add-member-team_12').click();

  // Click the button
  await page.locator('#add-member-modal .actions button.ok').click();

  // Verify that the user was added
  await expect(page.locator('.organization.members .list a').getByText('user5 (User Five)')).toBeVisible();

  // Revert for repeatability
  const removeButton = page.locator('.delete-button[data-url="/org/org3/members/action/remove"][data-datauid="5"]');
  await expect(async () => {
    await removeButton.click();
    // A confirmation modal will appear
    await expect(page.locator('.modal#remove-organization-member')).toBeVisible();
    // Proceed removing from the org
    await page.locator('.modal#remove-organization-member .actions button.ok').click();
    await expect(page.locator('.organization.members .list a').getByText('user5 (User Five)')).toBeHidden();
  }).toPass();
});
