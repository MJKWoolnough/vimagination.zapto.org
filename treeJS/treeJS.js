"use strict";
window.addEventListener("load", function() {
	var rpc = new (function(onload){
		var ws = new WebSocket("ws://" + window.location.host + "/FH/rpc"),
		    requests = [],
		    nextID = 0,
		    request = function (method, params, callback) {
			    var msg = {
				    "method": "RPC." + method,
				    "id": nextID,
				    "params": [params],
			    };
			    requests[nextID] = callback;
			    ws.send(JSON.stringify(msg));
			    nextID++;
		    },
		    closed = false;
		ws.onmessage = function (event) {
			var data = JSON.parse(event.data),
			req = requests[data.id];
			delete requests[data.id];
			if (typeof req === "undefined") {
				return;
			} else if (data.error !== null) {
				alert(data.error);
				return;
			}
			req(data.result);
		};
		ws.onopen = onload;
		ws.onerror = function(event) {
			document.body.setInnerText("An error occurred");
		}
		ws.onclose = function(event) {
			if (closed === true) {
				return;
			}
			switch(event.code) {
			case 1006:
				document.body.setInnerText("The server unexpectedly closed the connection - this may be an error.");
				break;
			case 4000:
				document.body.setInnerText("The server closed the connection due to another session opening.");
				break;
			default:
				document.body.setInnerText("Lost Connection To Server! Code: " + event.code);
			}
		}
		window.addEventListener("beforeunload", function() {
			closed = true;
			ws.close();
		});
		this.getPerson = request.bind(this, "GetPerson");
		this.getFamily = request.bind(this, "GetFamily");
	})(function(){
		init();
	}),
	genderMale = 77,
	genderFemale = 70,
	genderUnknown = 85,
	createElement = (function(){
		var ns = document.getElementsByTagName("html")[0].namespaceURI;
		return document.createElementNS.bind(document, ns);
	}()),
	waitGroup = function(callback) {
		var state = 0;
		this.add = function(amount) {
			amount = amount || 1;
			state += amount;
		};
		this.done = function() {
			state--;
			if (state === 0) {
				callback();
			}
		};
	},
	Person = (function() {
		var obj = function(data) {
			this.ID = data.ID;
			this.Firstname = data.FirstName;
			this.Surname = data.Surname;
			this.DOB = data.DOB;
			this.DOD = data.DOD;
			this.Gender = data.Gender;
			this.ChildOf = data.ChildOf;
			this.SpouseOf = data.SpouseOf;
			this.Expand = false;
		};
		obj.prototype = {
			getSpouses: function() {
				var spouses = new Array(this.SpouseOf.length);
				for (var i = 0; i < this.SpouseOf.length; i++) {
					spouses[i] = cache.getFamily(this.SpouseOf[i]);
				}
				return spouses;
			},
			getChildOf: function() {
				return cache.getFamily(this.ChildOf);
			}
		};
		return obj;
	}()),
	Family = (function() {
		var obj = function(data) {
			this.ID = data.ID;
			this.Husband = data.Husband;
			this.Wife = data.Wife;
			this.Children = data.Children;
		};
		obj.prototype = {
			getHusband: function() {
				return cache.getPerson(this.Husband);
			},
			getWife: function() {
				return cache.getPerson(this.Wife);
			},
			getChildren: function() {
				var children = new Array(this.Children.length);
				for (var i = 0; i < this.Children.length; i++) {
					children[i] = cache.getPerson(this.Children[i]);
				}
				return children;
			}
		};
		return obj;
	}()),
	cache = new (function() {
		var unknownPerson = new Person({
			ID: 0,
			FirstName: "?",
			Surname: "?",
			DOB: "",
			DOD: "",
			Gender: genderUnknown,
			ChildOf: 0,
			SpouseOf: []
		    }),
		    unknownFamily = new Family({
			ID: 0,
			Husband: 0,
			Wife: 0,
			Children: []
		    }),
		    personCache = [unknownPerson],
		    familyCache = [unknownFamily];
		this.getPerson = function(id) {
			var pc = personCache[id];
			if (typeof pc !== "undefined") {
				return pc;
			}
			return unknownPerson;
		};
		this.getFamily = function(id) {
			var fc = familyCache[id]
			if (typeof fc !== "undefined") {
				return fc;
			}
			return unknownFamily;
		};
		this.loadPerson = function(id, callback) {
			var pc = personCache[id];
			if (typeof pc !== "undefined") {
				callback(pc);
				return;
			}
			rpc.getPerson(id, function(person) {
				var p = new Person(person);
				personCache[id] = p;
				callback(p);
			});
		};
		this.loadFamily = function(id, callback) {
			var fc = familyCache[id];
			if (typeof fc !== "undefined") {
				callback(fc);
				return;
			}
			rpc.getFamily(id, function(family) {
				var f = new Family(family);
				familyCache[id] = f;
				callback(f);
			});
		};
		this.expandPerson = function(id, callback) {
			cache.loadPerson(id, function(person) {
				var wg = new waitGroup(callback),
				    wgDone = wg.done.bind(wg),
				    familyLoader = function(family) {
					if (family.Husband != 0) {
						wg.add(1);
						cache.loadPerson(family.Husband, wgDone);
					}
					if (family.Wife != 0) {
						wg.add(1);
						cache.loadPerson(family.Wife, wgDone);
					}
					if (family.Children.length > 0) {
						wg.add(family.Children.length);
						for(var i = 0; i < family.Children.length; i++) {
							cache.loadPerson(family.Children[i], wgDone);
						}
					}
					wg.done();
				};
				wg.add(1 + person.SpouseOf.length);
				cache.loadFamily(person.ChildOf, familyLoader);
				for (var i = 0; i < person.SpouseOf.length; i++) {
					cache.loadFamily(person.SpouseOf[i], familyLoader);
				}
			});
		}
	})(),
	focusID = 0,
	highlight = [],
	lines, boxes,
	rowStart = 100,
	colStart = 50,
	rowGap = 150,
	colGap = 200,
	boxWidth = 150,
	chosenX = 0,
	chosenY = 0,
	gender = function(g) {
		if (g === genderMale) {
			return "M";
		} else if (g === genderFemale) {
			return "F";
		}
		return "U";
	},
	personBox = function(person, row, col, spouse) {
		var d = createElement("div"),
		    style = d.style,
		    y = rowStart + row * rowGap,
		    x = colStart + col * colGap,
		    className = "person sex_" + gender(person.Gender),
		    name = createElement("div");
		style.setProperty("top", y + "px");
		style.setProperty("left", x + "px");
		if (person.ID === focusID) {
			d.id = "chosen";
			chosenX = x;
			chosenY = y;
		}
		for (var i = 0; i < highlight.length; i++) {
			if (person.ID === highlight[i]) {
				className += " highlight";
			}
		}
		if (person.SpouseOf.length > 0) {
			var collapseExpand = createElement("div");
			if (!person.Expand || spouse) {
				collapseExpand.className = "expand";
			} else {
				collapseExpand.className = "collapse";
			}
			d.appendChild(collapseExpand);
			d.addEventListener("click", expandCollapse(person, !person.Expand, spouse));
			className += " clicky";
		}
		name.className = "name";
		name.setInnerText(person.Firstname + " " + person.Surname);
		d.appendChild(name);
		if (person.DOB !== "") {
			var dob = createElement("div");
			dob.setInnerText("DOB: " + person.DOB);
			dob.className = "dob";
			d.appendChild(dob);
		}
		if (person.DOD !== "") {
			var dod = createElement("div");
			dod.setInnerText("DOD: " + person.DOD);
			dod.className = "dod";
			d.appendChild(dod);
		}
		d.className = className;
		boxes.appendChild(d);
	},
	expandCollapse = function(person, expand, spouse) {
		if (spouse) {
			return function() {
				focusID = person.ID;
				person.Expand = true;
				cache.expandPerson(person.ID, drawAndMove);
			};
		} else if (expand) {
			return function() {
				person.Expand = true;
				cache.expandPerson(person.ID, justDraw);
			};
		}
		return function() {
			person.Expand = false;
			justDraw();
		};
	},
	marriage = function(row, start, end) {
		var d = createElement("div"),
		    style = d.style;
		d.className = "spouseLine";
		style.setProperty("top", (rowStart + row * rowGap) + "px");
		style.setProperty("left", (colStart + start * colGap) + "px");
		style.setProperty("width", ((end - start) * colGap) + "px");
		lines.appendChild(d);
	},
	downLeft = function(row, start, end) {
		var d = createElement("div"),
		    style = d.style;
		d.className = "downLeft";
		style.setProperty("top", (rowStart + row * rowGap) + "px");
		style.setProperty("left", (colStart + start * colGap - 125) + "px");
		style.setProperty("width", ((end - start) * colGap + 100) + "px");
		lines.appendChild(d);
	},
	downRight = function(row, col) {
		var d = createElement("div"),
		    style = d.style;
		d.className = "downRight";
		style.setProperty("top", (rowStart + row * rowGap) + "px");
		style.setProperty("left", (colStart + col * colGap - 25) + "px");
		lines.appendChild(d);
	},
	siblingUp = function(row, col) {
		var d = createElement("div"),
		    style = d.style;
		d.className = "downLeft";
		style.setProperty("top", (rowStart + row * rowGap - 50) + "px");
		style.setProperty("left", (colStart + col * colGap + 75) + "px");
		style.setProperty("width", "0");
		style.setProperty("height", "50px");
		lines.appendChild(d);
	},
	siblingLine = function(row, start, end) {
		var d = createElement("div"),
		    style = d.style;
		d.className = "downLeft";
		style.setProperty("top", (rowStart + row * rowGap - 50) + "px");
		style.setProperty("left", (colStart + start * colGap + 75) + "px");
		style.setProperty("width", ((end - start) * colGap) + "px");
		style.setProperty("height", "0");
		lines.appendChild(d);
	},
	rows = new (function() {
		var rows = new Array();
		this.getRow = function(row) {
			var r = rows[row];
			if (typeof r === "undefined") {
				return 0;
			}
			return r;
		};
		this.setRow = function(row, col) {
			rows[row] = col;
		};
		this.rowPP = function(row) {
			var r = this.getRow(row);
			this.setRow(row, r + 1);
			return r;
		};
		this.reset = function() {
			rows = new Array();
		};
	})(),
	Box = (function() {
		var obj = function(row) {
			this.row = row;
			this.col = rows.rowPP(row);
		};
		obj.prototype = {
			setCol: function(col) {
				this.col = col;
				if (col >= rows.getRow(this.row)) {
					rows.setRow(this.row, col + 1);
				}
			},
			addCol: function(col) {
				this.setCol(this.col + col);
			}
		};
		return obj;
	}()),
	Children = (function() {
		var obj = function(family, parents, row) {
			var children = family.getChildren();
			this.parents = parents;
			this.children = new Array(children.length);
			for (var i = 0; i < children.length; i++) {
				this.children[i] = new Child(children[i], row);
			}
			if (typeof parents.spouses !== "undefined") {
				this.shift(0);
			}
		};
		obj.prototype = {
			shift: function(diff) {
				if (this.children.length > 0) {
					var pDiff = this.parents.box.col + diff - 1 - this.children[this.children.length-1].lastX();
					if (pDiff > 0) {
						for (var i = this.children.length - 1; i >= 0; i--) {
							if (!this.children[i].shift(pDiff)) {
								return false;
							}
						}
					}
				}
				return true;
			},
			draw: function() {
				if (this.children.length > 1) {
					siblingLine(this.children[0].box.row, this.children[0].box.col, this.children[this.children.length-1].box.col);
				}
				for (var i = 0; i < this.children.length; i++) {
					this.children[i].draw();
				}
			}
		};
		return obj;
	}()),
	Child = (function() {
		var obj = function(person, row) {
			this.person = person;
			this.box = new Box(row);
			if (person.Expand) {
				this.spouses = new Spouses(person.getSpouses(), this, row);
			} else {
				this.spouses = {spouses: []};
			}
		};
		obj.prototype = {
			lastX: function() {
				if (this.spouses.spouses.length > 0) {
					return this.spouses.spouses[this.spouses.spouses.length - 1].box.col;
				}
				return this.box.col;
			},
			shift: function(diff) {
				if (this.spouses.spouses.length > 0) {
					if (!this.spouses.shift(diff)) {
						return false;
					}
					this.box.setCol(this.spouses.spouses[0].box.col - 1);
				} else {
					this.box.addCol(diff);
				}
				return true;
			},
			draw: function() {
				siblingUp(this.box.row, this.box.col);
				personBox(this.person, this.box.row, this.box.col, false);
				if (this.spouses.spouses.length > 0) {
					this.spouses.draw()
				}
			}
		};
		return obj;
	}()),
	Spouses = (function() {
		var obj = function(families, spouse, row) {
			this.spouse = spouse;
			this.spouses = new Array(families.length);
			for (var i = 0; i < families.length; i++) {
				if (spouse.person.Gender === genderFemale) {
					this.spouses[i] = new Spouse(families[i], families[i].getHusband(), this, row);
				} else {
					this.spouses[i] = new Spouse(families[i], families[i].getWife(), this, row);
				}
			}
			if (families.length > 0) {
				spouse.box.col = this.spouses[0].box.col - 1;
			}
		};
		obj.prototype = {
			shift: function(diff) {
				for (var i = this.spouses.length - 1; i >= 0; i--) {
					if (!this.spouses[i].shift(diff)) {
						return false;
					}
				}
				return true;
			},
			draw: function() {
				if (this.spouses.length > 0) {
					marriage(this.spouse.box.row, this.spouse.box.col, this.spouses[this.spouses.length - 1].box.col);
					for (var i = 0; i < this.spouses.length; i++) {
						this.spouses[i].draw();
					}
				}
			}
		};
		return obj;
	}()),
	Spouse = (function() {
		var obj = function(family, person, spouses, row) {
			this.spouses = spouses;
			this.person = person;
			this.box = new Box(row);
			this.children = new Children(family, this, row + 1);
			if (family.Children.length > 0) {
				var firstChildPos = this.children.children[0].box.col;
				if (this.box.col < firstChildPos) {
					this.box.setCol(firstChildPos);
				}
			}
		};
		obj.prototype = {
			shift: function(diff) {
				var all = true;
				if (this.children.children.length > 0) {
					all = this.children.shift(diff);
				}
				this.box.addCol(diff);
				return all;
			},
			draw: function() {
				personBox(this.person, this.box.row, this.box.col, true);
				if (this.children.children.length > 0) {
					if (this.box.col === this.children.children[0].box.col) {
						downRight(this.box.row, this.box.col);
					} else if (this.box.col > this.children.children[this.children.children.length - 1].box.col) {
						downLeft(this.box.row, this.children.children[this.children.children.length - 1].box.col + 1, this.box.col);
					} else {
						downLeft(this.box.row, this.box.col, this.box.col);
					}
					this.children.draw();
				}
			}
		};
		return obj;
	}()),
	drawTree = function(move) {
		var top = cache.getPerson(focusID),
		    topFam, spouse;
		while(true) {
			var f = top.getChildOf(),
			    h = f.getHusband(),
			    w = f.getWife();
			if (h.Expand) {
				top = h;
			} else if (w.Expand) {
				top = w;
			} else {
				break;
			}
		}
		topFam = top.getChildOf();
		if (topFam.ID === 0) {
			topFam = new Family({Children: [top.ID]});
		}

		lines = document.createDocumentFragment();
		boxes = document.createDocumentFragment();

		personBox(topFam.getHusband(), 0, 0, false);
		personBox(topFam.getWife(), 0, 1, false);
		marriage(0, 0, 1);
		downLeft(0, 1, 1);

		rows.setRow(0, 1);
		spouse = {box: new Box(0)};
		(new Children(topFam, spouse, 1)).draw();

		
		document.body.removeChildren();
		document.body.appendChild(lines);
		document.body.appendChild(boxes);

		rows.reset();
		if (move) {
			window.scrollTo(chosenX - document.documentElement.clientWidth / 2, chosenY - document.documentElement.clientHeight / 2);
		}

	},
	drawAndMove = drawTree.bind(null, true),
	justDraw = drawTree.bind(null, false),
	init = function() {
		var search = window.location.search,
		    searches;
		if (search.length > 0 && search.charAt(0) === "?") {
			search = search.substr(1);
		}
		searches = search.split("&");
		for (var i = 0; i < searches.length; i++) {
			var keyValue = searches[i].split("=", 2);
			switch (keyValue[0]) {
			case "id":
				focusID = parseInt(keyValue[1], 10);
				break;
			case "highlight":
				var highlights = keyValue[1].split(",");
				for (var j = 0; j < highlights.length; j++) {
					highlight.push(parseInt(highlights[j]));
				}
				break;
			}
		}
		cache.expandPerson(focusID, function() {
			cache.getPerson(focusID).Expand = true;
			if (highlight.length > 0) {
				var wg = new waitGroup(function() {
					for (var i = 0; i < highlight.length; i++) {
						cache.getPerson(highlight[i]).Expand = true;
					}
					drawAndMove();
				}),
				    wgDone = wg.done.bind(wg);
				wg.add(highlight.length + 1);
				for (var i = 0; i < highlight.length; i++) {
					cache.expandPerson(highlight[i], wgDone);
				}
				wg.done();
			} else {
				drawTree(true);
			}
		});
	};
	Element.prototype.removeChildren = (function() {
		var docFrag = document.createDocumentFragment();
		return function(filter) {
			if (typeof filter === "function") {
				while (this.hasChildNodes()) {
					if (filter(this.firstChild)) {
						this.removeChild(this.firstChild);
					} else {
						docFrag.appendChild(this.firstChild);
					}
				}
				this.appendChild(docFrag);
			} else {
				while (this.hasChildNodes()) {
					this.removeChild(this.lastChild);
				}
			}
		};
	}());
	Element.prototype.setInnerText = function(text) {
		this.removeChildren();
		this.appendChild(document.createTextNode(text));
		return this;
	};
});
