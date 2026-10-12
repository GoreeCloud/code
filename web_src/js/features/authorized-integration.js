import $ from 'jquery';
import {createCodemirror} from './codemirror.ts';

export function initAuthorizedIntegrationClaimRuleEditor() {
  if (!$('.user.authorized-integrations #claim_rules').length) return;
  const _promise = createCodemirror($('#claim_rules')[0], 'claims.json', {language: 'JSON'});
}
