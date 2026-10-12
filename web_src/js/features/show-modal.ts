// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

import {showModal} from '../modules/modal.ts';

// Initialize all elements that have the `show-modal` class. The modal ID that
// is specified in the `data-modal` attribute will be shown. The shown modal
// can be modified by adding more attributes:
// * `data-modal-$TARGET="$VALUE"`, If $TARGET contains a dot then its split
// as $TARGET and $ATTR. $TARGET will first be queried as an identifier, then as
// a classname and then as an element tag name in the modal element. If $ATTR
// exists then the target element will have attribute $ATTR set to value $VALUE,
// otherwise if the element is of type input or textarea then the value is set
// to $VALUE otherwise the textContent of that element is set to $VALUE.
export function initGlobalShowModal() {
  document.addEventListener('click', (e) => {
    if (!(e.target instanceof Element)) {
      return;
    }
    const target = e.target.closest('.show-modal');
    if (!target) {
      return;
    }
    e.preventDefault();

    const modal = document.querySelector<HTMLDialogElement>(target.getAttribute('data-modal'));
    if (!modal) {
      throw new Error('No modal found for this action');
    }

    const modalAttrPrefix = 'data-modal-';
    for (const attrib of (target as HTMLElement).attributes) {
      if (!attrib.name.startsWith(modalAttrPrefix)) {
        continue;
      }

      const attrTargetCombo = attrib.name.substring(modalAttrPrefix.length);
      const [attrTargetName, attrTargetAttr] = attrTargetCombo.split('.');

      // try to find target by: "#target" -> ".target" -> "target tag"
      const attrTarget = modal.querySelector(`#${attrTargetName}, .${attrTargetName}, ${attrTargetName}`);

      if (attrTargetAttr) {
        attrTarget.setAttribute(attrTargetAttr, attrib.value);
      } else if (attrTarget instanceof HTMLInputElement || attrTarget instanceof HTMLTextAreaElement) {
        attrTarget.value = attrib.value; // FIXME: add more supports like checkbox
      } else {
        attrTarget.textContent = attrib.value; // FIXME: it should be more strict here, only handle div/span/p
      }
    }

    showModal(modal, undefined);
  });
}
