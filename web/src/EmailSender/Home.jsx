import React from "react";
import Avatar from "@mui/material/Avatar";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardHeader from "@mui/material/CardHeader";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTheme } from "@mui/material/styles";
import { Mail } from "@mui/icons-material";

const inputs = [
  { label: "From" },
  { label: "Subject" },
  { label: "Content", multiline: true },
];

const EmailSender = () => {
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
      <Grid item lg={3}>
        <Card variant="outlined">
          <CardHeader
            avatar={
              <Avatar>
                <Mail />
              </Avatar>
            }
            title="Email"
            subheader="Send an email"
          />
          <CardContent>
            <Grid container spacing={2}>
              {inputs.map((input) => (
                <Grid item xs={12} key={input.label}>
                  <TextField
                    name={input.label}
                    label={input.label}
                    placeholder={input.label}
                    fullWidth
                    minRows={!!input.multiline ? 4 : 0}
                    multiline={!!input.multiline}
                  />
                </Grid>
              ))}
            </Grid>
          </CardContent>
          <CardActions>
            <Button variant="outlined">Send</Button>
          </CardActions>
        </Card>
      </Grid>
    </Grid>
  );
};

export default EmailSender;
