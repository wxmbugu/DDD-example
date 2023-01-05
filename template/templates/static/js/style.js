const toggle = document.querySelector('.toggle')
const menu = document.querySelector('.menu')

/* Toggle mobile menu */

function toggleMenu() {
  if (menu.classList.contains('active')) {
    menu.classList.remove('active')

    // adds the menu (hamburger) icon

    toggle.querySelector('a').innerHTML =
      '<img style="width: 48px;height: 48px; background-color: white;border-radius: 28px;"src="/static/templates/static/images/menu.png"alt="Hamburger Menu"/>'
  } else {
    menu.classList.add('active')

    // adds the close (x) icon

    toggle.querySelector('a').innerHTML = '<p>X</p>'
  }
}

/* Event Listener */

toggle.addEventListener('click', toggleMenu, false)

const items = document.querySelectorAll('.item')

/* Activate Submenu */

function toggleItem() {
  if (this.classList.contains('submenu-active')) {
    this.classList.remove('submenu-active')
  } else if (menu.querySelector('.submenu-active')) {
    menu.querySelector('.submenu-active').classList.remove('submenu-active')

    this.classList.add('submenu-active')
  } else {
    this.classList.add('submenu-active')
  }
}

/* Event Listeners */

for (let item of items) {
  if (item.querySelector('.submenu')) {
    item.addEventListener('click', toggleItem, false)

    item.addEventListener('keypress', toggleItem, false)
  }
}

/* Close Submenu From Anywhere */

function closeSubmenu(e) {
  if (menu.querySelector('.submenu-active')) {
    let isClickInside = menu

      .querySelector('.submenu-active')

      .contains(e.target)

    if (!isClickInside && menu.querySelector('.submenu-active')) {
      menu.querySelector('.submenu-active').classList.remove('submenu-active')
    }
  }
}

/* Event listener */
document.addEventListener('click', closeSubmenu, false)

let form = document.querySelecter('form')

form.addEventListener('submit', (e) => {
  e.preventDefault()
  return false
})
