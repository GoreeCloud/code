// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
import {showModal} from '../modules/modal.ts';

export function initAddOrgMemberButton() {
  const addMemberButton = document.querySelector('#add-org-member-button');
  addMemberButton?.addEventListener('click', () => {
    showModal('add-member-modal', () => {
      const form: HTMLFormElement = document.querySelector('.add-member.form');
      if (!form.checkValidity()) {
        form.reportValidity();
        return false;
      }
      const selectedTeamCount = document.querySelectorAll('.add-member.form input[type=checkbox]:checked').length;
      const noTeamSelectedMessage : HTMLElement = document.querySelector('#no-team-selected-message');
      noTeamSelectedMessage.hidden = (selectedTeamCount !== 0);
      if (selectedTeamCount === 0) {
        return false;
      }
      form.requestSubmit();
    });
    return false;
  });
}
