import React from "react";
import Autocomplete from "@mui/material/Autocomplete";
import Avatar from "@mui/material/Avatar";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardHeader from "@mui/material/CardHeader";
import Chip from "@mui/material/Chip";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTheme } from "@mui/material/styles";
import { Mail, Send, Delete } from "@mui/icons-material";

const inputs = [{ label: "Subject" }, { label: "Content", multiline: true }];

const EmailSender = () => {
  const theme = useTheme();
  const [recipients, setRecipients] = React.useState([]);

  const onRecipients = (event, value) => {
    setRecipients(value);
  };

  console.log(recipients);

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
                <Mail variant="rounded" />
              </Avatar>
            }
            title="Email"
            subheader="Send an email"
          />
          <CardContent>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <Autocomplete
                  multiple
                  freeSolo
                  id="recipients"
                  options={[]}
                  defaultValue={[]}
                  onChange={onRecipients}
                  renderTags={(value, getTagProps) =>
                    value.map((option, index) => (
                      <Chip
                        variant="outlined"
                        label={option}
                        {...getTagProps({ index })}
                      />
                    ))
                  }
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      variant="outlined"
                      label="Recipients"
                    />
                  )}
                />
              </Grid>
              {inputs.map((input) => (
                <Grid item xs={12} key={input.label}>
                  <TextField
                    name={input.label.toLowerCase()}
                    label={input.label}
                    fullWidth
                    minRows={!!input.multiline ? 4 : 0}
                    multiline={!!input.multiline}
                  />
                </Grid>
              ))}
            </Grid>
          </CardContent>
          <CardActions>
            <Button startIcon={<Send variant="rounded" />} variant="contained">
              Send
            </Button>
            <Button startIcon={<Delete variant="rounded" />} variant="outlined">
              Discard
            </Button>
          </CardActions>
        </Card>
      </Grid>
    </Grid>
  );
};

export default EmailSender;
