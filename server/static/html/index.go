package html

var Index = `
<html>
	<head>

		<title>Biorcell3D temperature probe viewer.</title>

		<!-- Required meta tags -->
		<meta charset="utf-8">
    	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

		<!-- MaterialDesign icons CSS -->
		<link href="/css/materialdesignicons.min.css" rel="stylesheet"></link>

		<!-- Bootstrap CSS -->
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@4.5.3/dist/css/bootstrap.min.css" integrity="sha384-TX8t27EcRE3e/ihU7zmQxVncDAy5uIKz4rEkgIXeMed4M0jlfIDPvg6uqKI2xXr2" crossorigin="anonymous">

		<!-- Web Assembly -->
		<script src="/js/wasm_exec.js"></script>
		<script>
			const go = new Go();
			WebAssembly.instantiateStreaming(fetch("/wasm/wasm"), go.importObject).then((result) => {
				go.run(result.instance);
			});
		</script>
	
		<style type="text/css">
		.highlighted {
			background-color: yellow;
		}
		</style>

	</head>
	<body>
		<div id="header" class="p-sm-2">
			<img class="d-none d-sm-block" width="300" alt="logo" src="{{.Base64Logo}}"/>
			<button title="global configuration" type="button" class="btn btn-secondary position-absolute" style="top: 2px; right: 2px;">
				<span class="mdi mdi-cog-outline"></span>
			</button>
		</div>
		<div id="content" class="container">
			<div id="temperatureRecord" class="row">

			</div>
		</div>
		<div id="footer" class="p-sm-2 fixed-bottom d-none d-sm-block">
				<span class="mdi mdi-copyright"></span>Biorcell3D 2020.
		</div>
	</body>

	<!-- jQuery and Bootstrap Bundle (includes Popper) -->
	<script src="https://code.jquery.com/jquery-3.5.1.slim.min.js" integrity="sha384-DfXdz2htPH0lsSSs5nCTpuj/zy4C+OGpamoFVy38MVBnE+IbbVYUew+OrCXaRkfj" crossorigin="anonymous"></script>
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@4.5.3/dist/js/bootstrap.bundle.min.js" integrity="sha384-ho+j7jyWK8fNQe+A12Hb8AhRq26LrZ/JpcUGGOn+Y7RsweNrtN/tE3MoK7ZeZDyx" crossorigin="anonymous"></script>	

    <script>

		// Highlight the element with the given id
		// for seconds.
		function Hightlight(id) {
			$(id).addClass("highlighted");
			setTimeout(function () {
				$(id).removeClass("highlighted");
			}, 2000);
		}

		// DOMContentLoaded is called by wasm
		// when loaded.
		function DOMContentLoaded() {

			let socket = new WebSocket("ws://{{.ServerAddress}}/ws");

			socket.onopen = () => {
				OnOpen();
			};

			socket.onclose = event => {
				OnClose();
			};

			socket.onerror = error => {
			};

			socket.onmessage = function(evt) {
				OnMessage(evt);
			};

		}
        
    </script>
</html>
`
