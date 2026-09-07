package email

const (
	BookingConfirmedEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Booking Confirmed</title>
</head>

<body style="margin:0; padding:0; background:#f5f7fa; font-family:Arial,Helvetica,sans-serif; color:#333;">
	<table width="100%" cellpadding="0" cellspacing="0" style="background:#f5f7fa; padding:40px 0;">
		<tr>
			<td align="center">

				<table width="600" cellpadding="0" cellspacing="0"
					style="max-width:600px; width:100%; background:#ffffff; border-radius:8px; overflow:hidden;">

					<tr>
						<td style="background:#16a34a; padding:24px 32px; color:#ffffff;">
							<h1 style="margin:0; font-size:24px;">
								Booking Confirmed
							</h1>
						</td>
					</tr>

					<tr>
						<td style="padding:32px;">

							<p style="font-size:16px;">
								Hi {{.Name}},
							</p>

							<p style="font-size:16px; line-height:1.6;">
								Great news! Your booking has been successfully confirmed.
							</p>

							<table width="100%" cellpadding="0" cellspacing="0"
								style="background:#f8fafc; border-radius:6px;">

								<tr>
									<td style="padding:16px;">
										<strong>Booking ID</strong>
									</td>

									<td align="right" style="padding:16px;">
										{{.BookingID}}
									</td>
								</tr>

							</table>

							<p style="font-size:15px; line-height:1.6;">
								Please keep your booking ID for future reference.
							</p>

						</td>
					</tr>

					<tr>
						<td style="padding:20px 32px; background:#f8fafc;
							color:#777; font-size:12px; text-align:center;">
							This is an automated notification. Please do not reply.
						</td>
					</tr>

				</table>

			</td>
		</tr>
	</table>
</body>
</html>
`

	ReservationConfirmedEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Reservation Confirmed</title>
</head>

<body style="margin:0; padding:0; background:#f5f7fa; font-family:Arial,Helvetica,sans-serif; color:#333;">
	<table width="100%" cellpadding="0" cellspacing="0" style="background:#f5f7fa; padding:40px 0;">
		<tr>
			<td align="center">

				<table width="600" cellpadding="0" cellspacing="0"
					style="max-width:600px; width:100%; background:#ffffff; border-radius:8px; overflow:hidden;">

					<tr>
						<td style="background:#2563eb; padding:24px 32px; color:#ffffff;">
							<h1 style="margin:0; font-size:24px;">
								Reservation Confirmed
							</h1>
						</td>
					</tr>

					<tr>
						<td style="padding:32px;">

							<p style="font-size:16px;">
								Hi {{.Name}},
							</p>

							<p style="font-size:16px; line-height:1.6;">
								Your reservation has been successfully confirmed.
							</p>

							<table width="100%" cellpadding="0" cellspacing="0"
								style="background:#f8fafc; border-radius:6px;">

								<tr>
									<td style="padding:14px 16px;">
										<strong>Booking ID</strong>
									</td>
									<td align="right" style="padding:14px 16px;">
										{{.BookingID}}
									</td>
								</tr>

								<tr>
									<td style="padding:14px 16px;">
										<strong>Reservation ID</strong>
									</td>
									<td align="right" style="padding:14px 16px;">
										{{.ReservationID}}
									</td>
								</tr>

							</table>

							<p style="font-size:15px; line-height:1.6;">
								Please keep your reservation ID for future reference.
							</p>

						</td>
					</tr>

					<tr>
						<td style="padding:20px 32px; background:#f8fafc;
							color:#777; font-size:12px; text-align:center;">
							This is an automated notification. Please do not reply.
						</td>
					</tr>

				</table>

			</td>
		</tr>
	</table>
</body>
</html>
`

	PaymentCompletedEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Payment Successful</title>
</head>

<body style="margin:0; padding:0; background:#f5f7fa; font-family:Arial,Helvetica,sans-serif; color:#333;">
	<table width="100%" cellpadding="0" cellspacing="0" style="background:#f5f7fa; padding:40px 0;">
		<tr>
			<td align="center">

				<table width="600" cellpadding="0" cellspacing="0"
					style="max-width:600px; width:100%; background:#ffffff; border-radius:8px; overflow:hidden;">

					<tr>
						<td style="background:#16a34a; padding:24px 32px; color:#ffffff;">
							<h1 style="margin:0; font-size:24px;">
								Payment Successful
							</h1>
						</td>
					</tr>

					<tr>
						<td style="padding:32px;">

							<p style="font-size:16px;">
								Hi {{.Name}},
							</p>

							<p style="font-size:16px; line-height:1.6;">
								Your payment has been successfully completed.
							</p>

							<table width="100%" cellpadding="0" cellspacing="0"
								style="background:#f8fafc; border-radius:6px;">

								<tr>
									<td style="padding:14px 16px;">
										<strong>Booking ID</strong>
									</td>
									<td align="right" style="padding:14px 16px;">
										{{.BookingID}}
									</td>
								</tr>

								<tr>
									<td style="padding:14px 16px;">
										<strong>Payment ID</strong>
									</td>
									<td align="right" style="padding:14px 16px;">
										{{.PaymentID}}
									</td>
								</tr>

								<tr>
									<td style="padding:14px 16px;">
										<strong>Total Fare</strong>
									</td>
									<td align="right"
										style="padding:14px 16px; font-size:18px;">
										₹{{.TotalFare}}
									</td>
								</tr>

							</table>

							<p style="font-size:15px; line-height:1.6;">
								Your payment has been recorded successfully.
								Please keep your payment ID for future reference.
							</p>

						</td>
					</tr>

					<tr>
						<td style="padding:20px 32px; background:#f8fafc;
							color:#777; font-size:12px; text-align:center;">
							This is an automated notification. Please do not reply.
						</td>
					</tr>

				</table>

			</td>
		</tr>
	</table>
</body>
</html>
`

	PaymentFailedEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Payment Failed</title>
</head>

<body style="margin:0; padding:0; background:#f5f7fa; font-family:Arial,Helvetica,sans-serif; color:#333;">
	<table width="100%" cellpadding="0" cellspacing="0" style="background:#f5f7fa; padding:40px 0;">
		<tr>
			<td align="center">

				<table width="600" cellpadding="0" cellspacing="0"
					style="max-width:600px; width:100%; background:#ffffff; border-radius:8px; overflow:hidden;">

					<tr>
						<td style="background:#dc2626; padding:24px 32px; color:#ffffff;">
							<h1 style="margin:0; font-size:24px;">
								Payment Failed
							</h1>
						</td>
					</tr>

					<tr>
						<td style="padding:32px;">

							<p style="font-size:16px;">
								Hi {{.Name}},
							</p>

							<p style="font-size:16px; line-height:1.6;">
								Unfortunately, we were unable to complete your payment.
							</p>

							<table width="100%" cellpadding="0" cellspacing="0"
								style="background:#fef2f2; border-radius:6px;">

								<tr>
									<td style="padding:14px 16px;">
										<strong>Booking ID</strong>
									</td>
									<td align="right" style="padding:14px 16px;">
										{{.BookingID}}
									</td>
								</tr>

								<tr>
									<td style="padding:14px 16px;">
										<strong>Payment ID</strong>
									</td>
									<td align="right" style="padding:14px 16px;">
										{{.PaymentID}}
									</td>
								</tr>

							</table>

							<p style="font-size:15px; line-height:1.6;">
								Please try the payment again. If the amount was deducted
								from your account, please contact support.
							</p>

						</td>
					</tr>

					<tr>
						<td style="padding:20px 32px; background:#f8fafc;
							color:#777; font-size:12px; text-align:center;">
							This is an automated notification. Please do not reply.
						</td>
					</tr>

				</table>

			</td>
		</tr>
	</table>
</body>
</html>
`
)
