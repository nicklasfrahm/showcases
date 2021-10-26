import React from "react";
import { Link } from "react-router-dom";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardMedia from "@mui/material/CardMedia";
import Grid from "@mui/material/Grid";
import Typography from "@mui/material/Typography";

const projects = [
  {
    title: "Email Sender",
    description:
      "Send emails and simulate automatic failover in case of an external service provider failure.",
    company: "Dreamdata",
    logo: "https://images.squarespace-cdn.com/content/v1/60880c8985e48a388d33bd16/1620732746386-HBC3RQJ55P25GMJ9PUY6/dreamdata.png",
    backgroundColor: "#b8dae5",
  },
];

const Dashboard = () => (
  <Grid
    container
    spacing={2}
    justifyContent="center"
    alignItems="center"
    sx={{ height: "100vh", width: "100vw" }}
  >
    {projects.map((p) => (
      <Grid md={3} item key={p.title}>
        <Card variant="outlined">
          <CardMedia>
            <Grid
              container
              spacing={2}
              justifyContent="center"
              alignItems="center"
              sx={{ height: "200px", backgroundColor: p.backgroundColor }}
            >
              <img style={{ width: "80%" }} src={p.logo} alt={p.title} />
            </Grid>
          </CardMedia>
          <CardContent>
            <Typography gutterBottom variant="h5" component="div">
              {p.title}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {p.description}
            </Typography>
          </CardContent>
          <CardActions>
            {/* TODO: Connect this to the authentication provider. */}
            <Button
              variant="contained"
              color="primary"
              to={p.title.toLowerCase().replace(/\s/g, "-")}
              component={Link}
              disabled={false}
            >
              View
            </Button>
          </CardActions>
        </Card>
      </Grid>
    ))}
  </Grid>
);

export default Dashboard;
