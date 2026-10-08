// @watch start
// templates/user/settings/**.tmpl
// web_src/css/{form,user}.css
// @watch end

import {expect} from '@playwright/test';
import {test, login_user, login} from './utils_e2e.ts';
import {screenshot} from './shared/screenshots.ts';
import {validate_form} from './shared/forms.ts';

test.beforeAll(async ({browser}, workerInfo) => {
  await login_user(browser, workerInfo, 'user2');
});

test('User: Profile settings', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings');

  await page.getByLabel('Full name').fill('SecondUser');

  const pronounsInput = page.locator('input[list="pronouns"]');
  await expect(pronounsInput).toHaveAttribute('placeholder', 'Unspecified');
  await pronounsInput.click();
  const pronounsList = page.locator('datalist#pronouns');
  const pronounsOptions = pronounsList.locator('option');
  const pronounsValues = await pronounsOptions.evaluateAll((opts) => opts.map((opt: HTMLOptionElement) => opt.value));
  expect(pronounsValues).toEqual(['he/him', 'she/her', 'they/them', 'it/its', 'any pronouns']);
  await pronounsInput.fill('she/her');

  await page.getByPlaceholder('Tell others a little bit').fill('I am a playwright test running for several seconds.');
  await page.getByPlaceholder('Tell others a little bit').press('Tab');
  await page.getByLabel('Website').fill('https://forgejo.org');
  await page.getByPlaceholder('Share your approximate').fill('on a computer chip');
  await page.getByLabel('User visibility').click();
  await page.getByLabel('Visible only to signed-in').click();
  await page.getByLabel('Hide email address Email address will').uncheck();
  await page.getByLabel('Hide activity from profile').check();

  await validate_form({page}, 'fieldset');
  await screenshot(page);
  await page.getByRole('button', {name: 'Update profile'}).click();
  await expect(page.getByText('Your profile has been updated.')).toBeVisible();
  await page.getByRole('link', {name: 'public activity'}).click();
  await expect(page.getByText('Your activity is only visible')).toBeVisible();
  await screenshot(page);

  await page.goto('/user2');
  await expect(page.getByText('SecondUser')).toBeVisible();
  await expect(page.getByText('on a computer chip')).toBeVisible();
  await expect(page.locator('li').filter({hasText: 'user2@example.com'})).toBeVisible();
  await expect(page.locator('li').filter({hasText: 'https://forgejo.org'})).toBeVisible();
  await expect(page.getByText('I am a playwright test')).toBeVisible();
  await screenshot(page);

  await page.goto('/user/settings');
  await page.locator('input[list="pronouns"]').fill('rob/ot');
  await page.getByLabel('User visibility').click();
  await page.getByLabel('Visible to everyone').click();
  await page.getByLabel('Hide email address Email address will').check();
  await page.getByLabel('Hide activity from profile').uncheck();
  await expect(page.getByText('Your profile has been updated.')).toBeHidden();
  await validate_form({page}, 'fieldset');
  await screenshot(page);
  await page.getByRole('button', {name: 'Update profile'}).click();
  await expect(page.getByText('Your profile has been updated.')).toBeVisible();

  await page.goto('/user2');
  await expect(page.getByText('SecondUser')).toBeVisible();
  await expect(page.locator('li').filter({hasText: 'user2@example.com'})).toBeHidden();
  await page.goto('/user2?tab=activity');
  await expect(page.getByText('Your activity is visible to everyone')).toBeVisible();
});

test('User: Storage overview', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings/storage_overview');
  await page.waitForLoadState();
  await page.getByLabel('Git LFS – 8 KiB').nth(1).hover({position: {x: 250, y: 2}});
  await expect(page.getByText('Git LFS – 8 KiB')).toBeVisible();

  // Show/hide legend by clicking on the bar
  await expect(page.locator('.stats ul').nth(1)).toBeHidden();
  await expect(page.getByText('Git LFS 8 KiB').nth(1)).toBeHidden();

  await page.locator('.stats summary').nth(1).click();
  await expect(page.locator('.stats ul').nth(1)).toBeVisible();
  await expect(page.getByText('Git LFS 8 KiB').nth(1)).toBeVisible();
  await screenshot(page);

  await page.locator('.stats summary').nth(1).click();
  await expect(page.locator('.stats ul').nth(1)).toBeHidden();
  await expect(page.getByText('Git LFS 8 KiB').nth(1)).toBeHidden();

  await screenshot(page);
});

test('User: Canceling adding SSH key clears inputs', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings/keys');
  await page.locator('#add-ssh-button').click();

  await page.getByLabel('Key name').fill('MyAwesomeKey');
  await page.locator('#ssh-key-content').fill('Wront key material');

  await page.getByRole('button', {name: 'Cancel'}).click();
  await page.locator('#add-ssh-button').click();

  const keyName = page.getByLabel('Key name');
  await expect(keyName).toHaveValue('');

  const content = page.locator('#ssh-key-content');
  await expect(content).toHaveValue('');
});

test('User: Canceling adding GPG key clears input', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings/keys');
  await page.locator('.show-panel[data-panel="#add-gpg-key-panel"]').click();

  const gpgKeyContent = page.locator('#gpg-key-content');
  await gpgKeyContent.fill('Wront key material');

  await page.locator('.hide-panel[data-panel="#add-gpg-key-panel"]').click();

  await expect(gpgKeyContent).toHaveValue('');
});

test('User: Add access token', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings/applications');
  await page.getByRole('link', {name: 'New access token'}).click();

  await page.locator('#scoped-access-submit').click();
  await page.locator('#name:invalid').isVisible();

  await page.selectOption('#access-token-scope-activitypub', 'read:activitypub');
  await page.locator('#scoped-access-submit').click();

  await page.locator('#name:invalid').isVisible();
  await expect(page.locator('#access-token-scope-activitypub')).toHaveValue('read:activitypub');

  const tokenName = globalThis.crypto.randomUUID();
  await page.locator('#name').fill(tokenName);
  await page.getByRole('radio', {name: /^All /}).click();
  await page.locator('#scoped-access-submit').click();

  await expect(page.locator('.ui.info.message.flash-info')).toBeVisible();
  const flashText = await page.locator('.ui.info.message.flash-info').textContent();
  expect(flashText?.trim()).toMatch(/^[0-9a-f]{40}$/);
  await page.getByText(tokenName).isVisible();
});

test('User: Add access token validation error', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings/applications');
  await page.getByRole('link', {name: 'New access token'}).click();

  await page.getByRole('button', {name: 'Generate token'}).click();
  await page.locator('#name:invalid').isVisible();

  await page.getByRole('textbox', {name: 'Token name *'}).fill('Token A');
  await page.getByRole('combobox', {name: 'activitypub'}).selectOption('read:activitypub');
  await page.getByRole('radio', {name: 'Public only'}).click();

  await page.getByRole('button', {name: 'Generate token'}).click();

  await page.getByText('has been used as an application name already.').isVisible();
  // validate that selected options (public-only, activitypub) are still selected after the validation error.
  await expect(page.getByRole('radio', {name: 'Public only'})).toBeChecked();
  await expect(page.getByRole('combobox', {name: 'activitypub'})).toHaveValue('read:activitypub');
});

test('User: Add specific repo access token', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings/applications');
  await page.getByRole('link', {name: 'New access token'}).click();

  const tokenName = globalThis.crypto.randomUUID();
  await page.getByRole('textbox', {name: /^Token name/}).fill(tokenName);
  await page.getByRole('combobox', {name: 'repository'}).selectOption('read:repository');

  // clicking specific repositories will display currently available repositories:
  await expect(page.getByText('org17/big_test_private_4')).toBeHidden();
  await page.getByRole('radio', {name: 'Specific repositories'}).click();
  await expect(page.getByText('org17/big_test_private_4')).toBeVisible();
  await expect(page.getByText('user2/commits_search_test')).toBeVisible(); // another repo, will be used to verify search worked

  await page.getByPlaceholder('Search repos…').fill('big_test_private_4');
  await page.getByRole('button', {name: 'Search…'}).click();

  // verify search results visible:
  await expect(page.getByText('org17/big_test_private_4')).toBeVisible();
  await expect(page.getByText('user2/commits_search_test')).toBeHidden();

  // after performing a search, verify that the token name, 'selected repositories', and selected permissions are maintained
  await expect(page.getByRole('textbox', {name: /^Token name/})).toHaveValue(tokenName);
  await expect(page.getByRole('radio', {name: 'Specific repositories'})).toBeChecked();
  await expect(page.getByRole('combobox', {name: 'repository'})).toHaveValue('read:repository');

  // Add the big_test_private_4 repo.
  await page.getByRole('button', {name: 'Add org17/big_test_private_4'}).click();
  await expect(page.getByText('Selected repository (1)')).toBeVisible();
  await expect(page.getByText('org17/big_test_private_4')).toBeVisible();

  // Remove it to test remove, and then re-add
  await page.getByRole('button', {name: 'Remove org17/big_test_private_4'}).click();
  await expect(page.getByText('Selected repositories (0)')).toBeVisible();
  await expect(page.getByText('org17/big_test_private_4')).toBeVisible();
  await page.getByRole('button', {name: 'Add org17/big_test_private_4'}).click();

  // Create the token and check for success.
  await page.getByRole('button', {name: 'Generate token'}).click();
  await expect(page.locator('.ui.info.message.flash-info')).toBeVisible();
  const flashText = await page.locator('.ui.info.message.flash-info').textContent();
  expect(flashText?.trim()).toMatch(/^[0-9a-f]{40}$/);
  await page.getByText(tokenName).isVisible();
});

// Test that validation errors on the repo-specific access token page retain all the entered field values when the
// error is displayed.
test('User: Add specific repo access token error', async ({browser}, workerInfo) => {
  const page = await login({browser}, workerInfo);
  await page.goto('/user/settings/applications');
  await page.getByRole('link', {name: 'New access token'}).click();

  await page.getByRole('textbox', {name: /^Token name/}).fill('Token A');
  await page.getByRole('combobox', {name: 'repository'}).selectOption('read:repository');
  await page.getByRole('radio', {name: 'Specific repositories'}).click();
  await page.getByRole('button', {name: 'Add org17/big_test_private_4'}).click();

  // Create the token, verify error, then check all the fields for retained values.
  await page.getByRole('button', {name: 'Generate token'}).click();
  await page.getByText('has been used as an application name already.').isVisible();

  await expect(page.getByRole('textbox', {name: /^Token name/})).toHaveValue('Token A');
  await expect(page.getByRole('radio', {name: 'Specific repositories'})).toBeChecked();
  await expect(page.getByRole('combobox', {name: 'repository'})).toHaveValue('read:repository');
  await expect(page.getByRole('button', {name: 'Remove org17/big_test_private_4'})).toBeVisible();
});
