// Copy-to-clipboard for every element with [data-copy].
document.querySelectorAll("[data-copy]").forEach((button) => {
  button.addEventListener("click", async () => {
    const text = button.getAttribute("data-copy");
    try {
      await navigator.clipboard.writeText(text);
      button.classList.add("copied");
      const label = button.querySelector(".copy-label");
      const previous = label ? label.textContent : null;
      if (label) label.textContent = "Copied!";
      setTimeout(() => {
        button.classList.remove("copied");
        if (label && previous !== null) label.textContent = previous;
      }, 1600);
    } catch {
      // Clipboard unavailable (permissions/http): select the text instead.
      const target = document.querySelector(button.getAttribute("data-copy-target") || "");
      if (target) {
        const range = document.createRange();
        range.selectNodeContents(target);
        const selection = window.getSelection();
        selection.removeAllRanges();
        selection.addRange(range);
      }
    }
  });
});

// Footer year.
const year = document.querySelector("#year");
if (year) year.textContent = String(new Date().getFullYear());
