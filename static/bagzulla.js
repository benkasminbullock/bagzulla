var leavebad = false;

function close_window() {
	if (leavebad) {
		return "Really?";
	}
	return false;
}

function mark_fixed() {
	document.getElementById('bug-status').getElementsByTagName('option')[1].selected = 'selected';
}

function setParts(partEl, text) {
	var parts = JSON.parse(text);
	var nParts = parts.length;
	var none = document.createElement("option");
	none.text = "None";
	none.value = 0;
	partEl.add(none);
	for (var i = 0; i < nParts; i++) {
		var option = document.createElement("option");
		option.text = parts[i].Name;
		option.value = parts[i].PartId;
		partEl.add(option);
	}
}

function removeOptions(partEl) {
	while (1) {
		var l = partEl.options.length;
		if (l == 0) {
			break;
		}
		partEl.remove(l - 1);
	}
}

function getParts(project) {
	var partEl = document.getElementById('part');
	removeOptions(partEl)
	var xhttp = new XMLHttpRequest();
	xhttp.onreadystatechange = function() {
		if (this.readyState == 4 && this.status == 200) {
			setParts(partEl, this.responseText);
		}
	};
	var url = topURL + "/project-parts/" + project;
	xhttp.open("GET", url, true);
	xhttp.send();
}

function removeParts() {
	var partEl = document.getElementById("part");
	partEl.style.display = "none";
}

function restoreParts() {
	var partEl = document.getElementById("part");
	partEl.style.display = "";
}

function setProject() {
	var project = document.getElementById("project").value;
	if (project == "13") {
		removeParts();
		return;
	}
	restoreParts();
	getParts(project);
}

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
