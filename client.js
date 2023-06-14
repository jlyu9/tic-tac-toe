let conn = new WebSocket("ws://127.0.0.1:8080/ws")  
// const conn = new WebSocket("ws://" + document.location.host + "/ws")
let symbol;

const board=document.querySelector('.board')
const cells = document.querySelectorAll('#cell')
conn.onmessage = function (evt) {
    console.log(evt)
    const data = JSON.parse(evt.data)
    // console.log(data)
    switch (data.tag) {
        case 'done':
           board.style.display = "grid"
            symbol=data.symbol
            if (symbol==1) {
                board.classList.add('circle')
            } else {
                board.classList.add('x')
            }
            if (symbol==2) {
                alert("You're first!")
            } else {
                alert("You're second!")
            }
            break
        case 'move':
            cells.forEach(cell => {
                if(!cell.classList.contains('x') && !cell.classList.contains('circle')){
                    cell.addEventListener('click', clickCell)
                }
            })
            break
        case 'update':
            if (data.symbol==1) {
                cells[data.index].classList.add('circle')
            } else {
                cells[data.index].classList.add('x')
            }
            break
    }
}

function clickCell(src){
    // var index=cells.indexOf(src.target)
    let i
    for (i=0;i<cells.length;i++) {
        if (cells[i]==src.target) {
            break
        }
    }
    if (symbol==1) {
        src.target.classList.add('circle')
    } else {
        src.target.classList.add('x')
    }
    cells.forEach(cell => cell.removeEventListener('click', clickCell))
    const payload={
        // 'tag':'moved',
        'index':i,
        'symbol':symbol
    }
    conn.send(JSON.stringify(payload))
}