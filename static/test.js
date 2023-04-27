"use strict";
(() => {
const modified_inputs = new Set;
const defaultValue = "defaultValue";
// store default values
addEventListener("beforeinput", (evt) => {
    const target = evt.target;
    if (!(defaultValue in target || defaultValue in target.dataset)) {
        target.dataset[defaultValue] = ("" + (target.value || target.textContent)).trim();
    }
});
// detect input modifications
addEventListener("input", (evt) => {
    const target = evt.target;
    let original;
    if (defaultValue in target) {
        original = target[defaultValue];
    } else {
        original = target.dataset[defaultValue];
    }
    if (original !== ("" + (target.value || target.textContent)).trim()) {
        if (!modified_inputs.has(target)) {
            modified_inputs.add(target);
        }
    } else if (modified_inputs.has(target)) {
        modified_inputs.delete(target);
    }
});
// clear modified inputs upon form submission
addEventListener("submit", () => {
    modified_inputs.clear();
    // to prevent the warning from happening, it is advisable
    // that you clear your form controls back to their default
    // state with form.reset() after submission
});
// warn before closing if any inputs are modified
addEventListener("beforeunload", (evt) => {
    if (modified_inputs.size) {
        const unsaved_changes_warning = "Changes you made may not be saved.";
        evt.returnValue = unsaved_changes_warning;
        return unsaved_changes_warning;
    }
});
})();
