// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// web_src/js/features/admin/**
// templates/admin/**
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';
import {screenshot} from './shared/screenshots.ts';

test.use({user: 'user1'});

test('Admin notices modal', async ({page}) => {
  const response = await page.goto('/admin/notices');
  expect(response?.status()).toBe(200);

  await page.getByText('description1').click();
  await expect(page.locator('#detail-modal .content')).toHaveText('description1');
  await screenshot(page, page.locator('#detail-modal'));
  await page.getByText('Cancel').click();
  await expect(page.locator('#change-email-modal')).toBeHidden();

  await page.getByText('description2').click();
  await expect(page.locator('#detail-modal .content')).toHaveText('description2');
  await screenshot(page, page.locator('#detail-modal'));
  await page.getByText('Cancel').click();
  await expect(page.locator('#change-email-modal')).toBeHidden();

  await page.getByText('description3').click();
  await expect(page.locator('#detail-modal .content')).toHaveText('description3');
  await screenshot(page, page.locator('#detail-modal'));
  await page.getByText('Cancel').click();
  await expect(page.locator('#change-email-modal')).toBeHidden();
});

test('Admin email list', async ({page}) => {
  const response = await page.goto('/admin/emails');
  expect(response?.status()).toBe(200);

  await page.locator('[data-uid="21"]').click();
  await expect(page.locator('#change-email-modal .content')).toHaveText('Are you sure you want to update this email address?');
  await screenshot(page, page.locator('#change-email-modal .content'));
  await page.locator('#email-action-form').getByText('No').click();
  await expect(page.locator('#change-email-modal')).toBeHidden();

  const activated = await page.locator('[data-uid="9"] .svg').evaluate((node) => node.classList.contains('octicon-check'));
  await page.locator('[data-uid="9"]').click();
  await page.getByRole('button', {name: 'Yes'}).click();

  // Retry-proof
  if (activated) {
    await expect(page.locator('[data-uid="9"] .svg')).toHaveClass(/octicon-x/);
  } else {
    await expect(page.locator('[data-uid="9"] svg')).toHaveClass(/octicon-check/);
  }
});
