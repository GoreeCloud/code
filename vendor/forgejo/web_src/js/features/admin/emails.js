import $ from 'jquery';
import {showModal} from '../../modules/modal.ts';

export function initAdminEmails() {
  function linkEmailAction(e) {
    const $this = $(this);
    $('#form-uid').val($this.data('uid'));
    $('#form-email').val($this.data('email'));
    $('#form-primary').val($this.data('primary'));
    $('#form-activate').val($this.data('activate'));
    showModal('change-email-modal', undefined);
    e.preventDefault();
  }
  $('.link-email-action').on('click', linkEmailAction);
}
