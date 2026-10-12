// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// templates/repo/branch/list.tmpl
// web_src/js/features/repo-branch.ts
// @watch end

import {expect} from '@playwright/test';
import {dynamic_id, test} from './utils_e2e.ts';
import {screenshot} from './shared/screenshots.ts';

test.use({user: 'user2'});

test('Create branch from branch', async ({page}) => {
  let response = await page.goto('/user2/repo1/branches');
  expect(response?.status()).toBe(200);

  // Open create branch modal.
  await page.getByRole('button', {name: 'Create new branch from "branch2"'}).click();
  await expect(page.locator('#create-branch-modal')).toBeVisible();
  await screenshot(page, page.locator('#create-branch-modal'));

  // Check that it can be cancelled.
  await page.getByRole('button', {name: 'Cancel'}).click();
  await expect(page.locator('#create-branch-modal')).toBeHidden();

  // Open it again and make a branch.
  await page.getByRole('button', {name: 'Create new branch from "branch2"'}).click();
  await expect(page.locator('#create-branch-modal')).toBeVisible();

  const branchName = dynamic_id();
  await page.getByRole('textbox').fill(branchName);
  await page.getByRole('button', {name: 'Confirm'}).click();

  // Verify branch exists.
  response = await page.goto(`/user2/repo1/src/branch/${branchName}`);
  expect(response?.status()).toBe(200);
});

test('Rename normal branch', async ({page}) => {
  let response = await page.goto('/user2/repo1/branches');
  expect(response?.status()).toBe(200);

  // Open rename branch modal.
  await page.locator('button[data-is-default-branch="false"]:not([data-old-branch-name="branch2"])').first().click();
  await expect(page.locator('#rename-branch-modal')).toBeVisible();
  await expect(page.locator('#rename-branch-modal .warning')).toBeHidden();
  await screenshot(page, page.locator('#rename-branch-modal'));

  // Check that it can be cancelled.
  await page.getByRole('button', {name: 'Cancel'}).click();
  await expect(page.locator('#rename-branch-modal')).toBeHidden();

  // Open it again and rename the branch.
  await page.locator('button[data-is-default-branch="false"]:not([data-old-branch-name="branch2"])').first().click();
  await expect(page.locator('#rename-branch-modal')).toBeVisible();
  await expect(page.locator('#rename-branch-modal .warning')).toBeHidden();

  const branchName = dynamic_id();
  await page.getByRole('textbox').fill(branchName);
  await page.getByRole('button', {name: 'Confirm'}).click();

  // Verify branch exists.
  response = await page.goto(`/user2/repo1/src/branch/${branchName}`);
  expect(response?.status()).toBe(200);
});

test('Rename default branch', async ({page}) => {
  let response = await page.goto('/user2/repo1/branches');
  expect(response?.status()).toBe(200);

  // Open rename branch modal.
  await page.locator('button[data-is-default-branch="true"]').click();
  await expect(page.locator('#rename-branch-modal')).toBeVisible();
  await expect(page.locator('#rename-branch-modal .warning')).toBeVisible();
  await screenshot(page, page.locator('#rename-branch-modal'));

  // Check that it can be cancelled.
  await page.getByRole('button', {name: 'Cancel'}).click();
  await expect(page.locator('#rename-branch-modal')).toBeHidden();

  // Open it again and rename the branch.
  await page.locator('button[data-is-default-branch="true"]').click();
  await expect(page.locator('#rename-branch-modal')).toBeVisible();
  await expect(page.locator('#rename-branch-modal .warning')).toBeVisible();

  const branchName = dynamic_id();
  await page.getByRole('textbox').fill(branchName);
  await page.getByRole('button', {name: 'Confirm'}).click();

  // Verify branch exists.
  response = await page.goto(`/user2/repo1/src/branch/${branchName}`);
  expect(response?.status()).toBe(200);
});
