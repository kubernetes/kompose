var icons = [
  'r6',
  'daytona',
  'zx6r',
  'gsxr',
  'cbr600rr'
]

var container = document.querySelector('.icons')

function spawn() {
  var icon = icons[Math.floor(Math.random() * icons.length)]

  var img = document.createElement("IMG")
  img.style.right = 25 + Math.random() * 50 + '%'
  img.setAttribute('class', 'icon')
  img.setAttribute("src", 'img/icons/' + icon + ".png")

  container.appendChild(img)

  setTimeout(function() {
    container.removeChild(img)
  }, 3000)
}

setTimeout(function run() {
  spawn()
  setTimeout(run, 1300 + Math.random() * 400)
}, 1300 + Math.random() * 400)

spawn()
