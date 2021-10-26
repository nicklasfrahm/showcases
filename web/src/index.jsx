import React from "react";
import ReactDOM from "react-dom";
import CssBaseline from "@mui/material/CssBaseline";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import { Switch, Redirect, Route, BrowserRouter } from "react-router-dom";
import reportWebVitals from "./reportWebVitals";
import Dashboard from "./Dashboard";
import EmailSender from "./EmailSender/Home";

const theme = createTheme({
  palette: {
    primary: {
      main: "#455a64",
    },
  },
});

ReactDOM.render(
  <React.StrictMode>
    <CssBaseline />
    <ThemeProvider theme={theme}>
      <BrowserRouter>
        <Switch>
          <Route exact path="/" component={Dashboard} />
          <Route exact path="/email-sender" component={EmailSender} />
          <Redirect to="/" />
        </Switch>
      </BrowserRouter>
    </ThemeProvider>
  </React.StrictMode>,
  document.getElementById("root")
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
