import React from "react";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardMedia from "@mui/material/CardMedia";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import Grid from "@mui/material/Grid";

const projects = [
  {
    title: "Email Sender Service",
    description:
      "Send emails and simulate automatic failover in case of an external service provider failure.",
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
        <Card>
          <CardMedia>
            <Grid
              container
              spacing={2}
              justifyContent="center"
              alignItems="center"
              sx={{ height: "150px", backgroundColor: p.backgroundColor }}
            >
              <img
                style={{ display: "block", width: "80%" }}
                src={p.logo}
                alt={p.title}
              />
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
            <Button color="primary">Login</Button>
          </CardActions>
        </Card>
      </Grid>
    ))}
  </Grid>
);
export default Dashboard;
