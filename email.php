<?php
if (isset($_POST['email']) && isset($_POST['recaptcha_response'])) {

  // Build POST request:
  $recaptcha_url = 'https://www.google.com/recaptcha/api/siteverify';
  $recaptcha_secret = getenv('RECAPTCHA_SECRET').PHP_EOL;;
  $recaptcha_response = $_POST['recaptcha_response'];

  // Make and decode POST request:
  $recaptcha = file_get_contents($recaptcha_url . '?secret=' . $recaptcha_secret . '&response=' . $recaptcha_response);
  $recaptcha = json_decode($recaptcha);

  // Take action based on the score returned:
  if ($recaptcha->score >= 0.5) {
    $errors = '';
    $myEmail = 'arthurva@arthurvardevanyan.com'; //<-----Put Your email address here.
    if (
      empty($_POST['name'])  ||
      empty($_POST['email']) ||
      empty($_POST['message'])
    ) {
      $errors .= "\n Error: all fields are required";
    }

    $name = $_POST['name'];
    $email_address = $_POST['email'];
    $message = $_POST['message'];

    if (!preg_match(
      "/^[_a-z0-9-]+(\.[_a-z0-9-]+)*@[a-z0-9-]+(\.[a-z0-9-]+)*(\.[a-z]{2,3})$/i",
      $email_address
    )) {
      $errors .= "\n Error: Invalid email address";
    }

    if (empty($errors)) {

      $to = $myEmail;

      $email_subject = "Contact form submission: $name";

      $email_body = "You have received a new message. " .

        " Here are the details:\n Name: $name \n " .

        "Email: $email_address\n Message \n $message";

      $headers = "From: $myEmail\n";

      $headers .= "Reply-To: $email_address";

      mail($to, $email_subject, $email_body, $headers);


      //Email Copy To Submitter
      //$headers .= "Reply-To: $myEmail";
      //$to = $email_address;
      //mail($to, $email_subject, $email_body, $headers);


      //redirect to the 'thank you' page

      //header('Location: index.html');

      echo '<script type="text/javascript">',
      'window.location="index.html#emailSent";',
      '</script>';
    }
  } else {
    // Not verified - show form error
    echo '<script type="text/javascript">',
    'window.location="index.html#emailFailed";',
    '</script>';
  }
}
