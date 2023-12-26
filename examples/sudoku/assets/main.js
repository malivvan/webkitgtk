class Sudoku {
    constructor(elem) {
        this.elem = elem;
        this.board = this.createBoard();
        this.controls = this.createControls();
        this.elem.appendChild(this.board)
        this.elem.appendChild(this.controls)
    }

    createControls() {
        let controls = document.createElement("div");
        controls.classList.add("controls");
        let genButton = this.createButton("Generate", this.fillBoard.bind(this));
        let genAmount = this.createInput("number", 20, 0, 81);
        controls.append(genButton, genAmount);
        return controls;
    }

    createButton(text, onClick) {
        let button = document.createElement("button");
        button.innerText = text;
        button.onclick = onClick;
        return button;
    }

    createInput(type, value, min, max) {
        let input = document.createElement("input");
        input.type = type;
        input.value = value;
        input.min = min;
        input.max = max;
        return input;
    }

    createBoard() {
        let board = document.createElement('div');
        board.className = 'board';
        for (let i = 0; i < 9; i++) {
            let row = document.createElement('div');
            row.className = 'row';
            for (let j = 0; j < 9; j++) {
                let cell = this.createCell(i, j);
                row.appendChild(cell);
            }
            board.appendChild(row);
        }
        return board;
    }

    createCell(i, j) {
        let cell = document.createElement('div');
        cell.className = 'cell';
        cell.id = `cell${i}${j}`;
        cell.innerHTML = ' ';
        cell.onclick = this.cellOnClick.bind(cell);
        cell.oncontextmenu = this.cellOnRightClick.bind(cell);
        return cell;
    }

    cellOnClick() {
        this.innerHTML = this.innerHTML === ' ' ? 1 : (parseInt(this.innerHTML) % 9) + 1;
    }

    cellOnRightClick(e) {
        e.preventDefault();
        this.innerHTML = this.innerHTML === ' ' ? 9 : (parseInt(this.innerHTML) - 1) || ' ';
    }

    getValues(selector) {
        let values = [];
        for (let i = 0; i < 9; i++) {
            values.push(parseInt(this.board.querySelector(selector(i)).innerHTML));
        }
        return values;
    }

    getRowValues = (row) => this.getValues(i => `#cell${row}${i}`);
    getColValues = (col) => this.getValues(i => `#cell${i}${col}`);

    getValidValues(row, col) {
        let rowValues = this.getRowValues(row);
        let colValues = this.getColValues(col);
        let squareValues = [1, 2, 3, 4, 5, 6, 7, 8, 9];
        return squareValues.filter(i => !rowValues.includes(i) && !colValues.includes(i));
    }

    getEmptyCells() {
        let emptyCells = [];
        for (let i = 0; i < 9; i++) {
            for (let j = 0; j < 9; j++) {
                if (this.board.querySelector(`#cell${i}${j}`).innerHTML === ' ') {
                    emptyCells.push([i, j]);
                }
            }
        }
        return emptyCells;
    }

    fillCell(row, col) {
        let cell = this.board.querySelector(`#cell${row}${col}`);

        let possibleValues = this.getValidValues(row, col);
        if (possibleValues.length === 0) {
            return false;
        }
        possibleValues.sort(() => Math.random() - 0.5);
        cell.classList.add("cell-generated");
        cell.innerHTML = possibleValues.pop();
        return true;
    }

    clearBoard() {
        for (let i = 0; i < 9; i++) {
            for (let j = 0; j < 9; j++) {
                let cell = this.board.querySelector(`#cell${i}${j}`);
                cell.classList.remove("cell-generated");
                cell.innerHTML = ' ';
            }
        }
    }

    fillBoard() {
        this.clearBoard();
        let emptyCells = this.getEmptyCells();
        emptyCells.sort(() => Math.random() - 0.5);
        let amount = parseInt(this.controls.querySelector("input").value);
        let n = 0;
        while (n < amount) {
            let [row, col] = emptyCells.pop();
            if (this.fillCell(row, col)) n++;
        }
    }
}

let game = new Sudoku(document.body);