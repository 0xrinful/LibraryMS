document.addEventListener("DOMContentLoaded", () => {
  const flashes = document.querySelectorAll("#flash-container .flash");

  flashes.forEach((flash) => {
    // Auto hide after 4s
    setTimeout(() => {
      flash.classList.add("hide");
      flash.addEventListener("animationend", () => flash.remove());
    }, 4000);

    // Optional manual close button
    const closeBtn = document.createElement("button");
    closeBtn.innerHTML = "&times;";
    closeBtn.classList.add("close");
    flash.appendChild(closeBtn);

    closeBtn.addEventListener("click", () => {
      flash.classList.add("hide");
      flash.addEventListener("animationend", () => flash.remove());
    });
  });
});
