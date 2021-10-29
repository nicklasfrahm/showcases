import React from "react";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardMedia from "@mui/material/CardMedia";
import Grid from "@mui/material/Grid";
import { useTheme } from "@mui/material/styles";
import Typography from "@mui/material/Typography";
import { Code } from "@mui/icons-material";

const baseURI = "https://github.com/nicklasfrahm/showcases";

const projects = [
  {
    title: "Email Sender",
    description:
      "Send emails and simulate automatic failover in case of an external service provider failure.",
    company: "Dreamdata",
    logo: "https://images.squarespace-cdn.com/content/v1/60880c8985e48a388d33bd16/1620732746386-HBC3RQJ55P25GMJ9PUY6/dreamdata.png",
    backgroundColor: "#b8dae5",
    documentation: baseURI,
  },
];

const Dashboard = () => {
  const theme = useTheme();

  return (
    <Grid
      container
      spacing={2}
      justifyContent="center"
      alignItems="center"
      sx={{
        height: "100vh",
        padding: theme.spacing(2),
      }}
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
              <Typography gutterBottom variant="h5" component="h2">
                {p.title}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {p.description}
              </Typography>
            </CardContent>
            <CardActions>
              {/* TODO: Connect this to the authentication provider. */}
              <Button
                variant="outlined"
                color="primary"
                href={p.documentation}
                disabled={false}
                startIcon={<Code variant="rouded" />}
              >
                View on GitHub
              </Button>
            </CardActions>
          </Card>
        </Grid>
      ))}
    </Grid>
  );
};

export default Dashboard;
