let page = 1;
let loading = false;

window.addEventListener("scroll", async () => {
  if (loading) return;

  if (window.innerHeight + window.scrollY >= document.body.offsetHeight - 100) {
    loading = true;
    page++;
    const res = await fetch(`/books?page=${page}`);
    const html = await res.text();
    document.querySelector(".books-grid").insertAdjacentHTML("beforeend", html);
    loading = false;
  }
});
