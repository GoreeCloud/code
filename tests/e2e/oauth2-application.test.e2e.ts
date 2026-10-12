// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// routers/web/auth/**
// templates/user/auth/**
// @watch end

import {expect} from '@playwright/test';
import {dynamic_id, test} from './utils_e2e.ts';

test.use({user: 'user2'});

test.describe('OAuth2 application', () => {
  const applicationName = dynamic_id();
  const state = dynamic_id();
  let redirectURI = '';
  let clientID = '';

  test('Create OAuth2 application', async ({page, baseURL}) => {
    const response = await page.goto('/user/settings/applications');
    expect(response?.status()).toBe(200);
    redirectURI = `${baseURL}/e2e-callback`;

    await page.getByRole('textbox', {name: 'Application name'}).fill(applicationName);
    await page.getByRole('textbox', {name: 'Redirect URIs.'}).fill(redirectURI);
    await page.getByRole('button', {name: 'Create application'}).click();

    await expect(page.locator('#flash-message')).toContainText('You have successfully created a new OAuth2 application.');

    clientID = (await page.getByRole('textbox', {name: 'Client ID'}).inputValue()).trim();
  });

  test('Authorize OAuth2 application', async ({page}) => {
    const response = await page.goto(`/login/oauth/authorize?client_id=${clientID}&redirect_uri=${encodeURIComponent(redirectURI)}&scope=read%3Auser&state=${state}&response_type=code`);
    expect(response?.status()).toBe(200);

    await expect(page.getByRole('heading')).toContainText(`Authorize "${applicationName}" to access your account?`);
    await expect(page.getByRole('main')).toContainText('With scopes: read:user.');
    await expect(page.getByRole('main')).toContainText(`You will be redirected to ${redirectURI} if you authorize this application.`);

    await page.getByRole('button', {name: 'Authorize Application'}).click();
    await page.waitForURL(new RegExp(`/e2e-callback\\?code=gta_[0-9a-z]{52}&state=${state}$`));
  });
});
