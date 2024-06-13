package cosmoauth

import (
	"fmt"
	"net/http"
	"text/template"
)

const redirectHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  <title>COSMO Auth redirector</title>
  <script type="module">
    const originalUrl = encodeURIComponent(window.location.href);
    const signInUrl = "{{ .SignInUrl }}" + "?redirect_to=" + originalUrl;
    window.location.href = signInUrl;
    console.log(signInUrl)
  </script>
  <style data-emotion="css" data-s="">
    @media (prefers-color-scheme: dark) {
      body {
        color: #fff;
        background-color: #121212;
      }
    }
    .root {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin: 80px;
    }
    .circularProgress {
      width: 40px;
      height: 40px;
      color: #673ab7;
      animation: animation-61bdi0 1.4s linear infinite;
    }
    @keyframes animation-61bdi0 {
      0% {
        transform: rotate(0deg);
      }
      100% {
        transform: rotate(360deg);
      }
    }
    .circularProgressSVG {
      display: block;
    }
    .circularProgressCircle {
      stroke: currentColor;
      stroke-dasharray: 80px, 200px;
      stroke-dashoffset: 0;
      animation: animation-1p2h4ri 1.4s ease-in-out infinite;
    }
    @keyframes animation-1p2h4ri {
      0% {
        stroke-dasharray: 1px, 200px;
        stroke-dashoffset: 0;
      }
      50% {
        stroke-dasharray: 100px, 200px;
        stroke-dashoffset: -15px;
      }
      100% {
        stroke-dasharray: 100px, 200px;
        stroke-dashoffset: -125px;
      }
    }
  </style>
</head>
<body>
  <div class="root">
    <span class="circularProgress">
      <svg class="circularProgressSVG" viewBox="22 22 44 44">
        <circle class="circularProgressCircle" cx="44" cy="44" r="20.2" fill="none" stroke-width="3.6" />
      </svg>
    </span>
    <p class="typography">redirect to signin page...</p>
  </div>
</body>
</html>
`

func writeRedirectHTML(w http.ResponseWriter, cfg *Config) error {
	t := template.New("redirect")
	t, err := t.Parse(redirectHTMLTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	err = t.Execute(w, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}

const forbiddenHTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>COSMO Auth</title>
    <style data-emotion="css" data-s="">
      @media (prefers-color-scheme: dark) {
        body {
          color: #fff;
          background-color: #121212;
        }
      }
      .root {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin: 80px;
      }
      .header {
        color: #e91e63;
      }
    </style>
  </head>
  <body>
    <div class="root">
      <h1 class="header">Forbidden</h1>
      <p class="typography">You are not allowed to access this page</p>
    </div>
  </body>
</html>
`

func writeForbiddenHTML(w http.ResponseWriter) {
	fmt.Fprint(w, forbiddenHTML)
}
