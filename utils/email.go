package utils

import (
    //"context"
    "fmt"
    "log"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ses"
)

const (
    Sender  = "no-reply@zetdisk.com" //change this to ur address
    CharSet = "UTF-8"
)

func SendOTPEmail(email, otp string) error {
    subject := "Your Password Reset OTP"
    body := fmt.Sprintf("Your OTP for password reset is: %s", otp)
    return SendEmail(email, subject, body)
}

func SendEmail(to, subject, body string) error {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("ap-south-1"), // change the aws resion 
    })
    if err != nil {
        log.Println("Failed to create AWS session:", err)
        return err
    }

    svc := ses.New(sess)

    input := &ses.SendEmailInput{
        Destination: &ses.Destination{
            ToAddresses: []*string{
                aws.String(to),
            },
        },
        Message: &ses.Message{
            Body: &ses.Body{
                Text: &ses.Content{
                    Charset: aws.String(CharSet),
                    Data:    aws.String(body),
                },
            },
            Subject: &ses.Content{
                Charset: aws.String(CharSet),
                Data:    aws.String(subject),
            },
        },
        Source: aws.String(Sender),
    }

    result, err := svc.SendEmail(input)
    if err != nil {
        log.Println("Failed to send email:", err)
        return err
    }

    log.Printf("Email sent to: %s\nSubject: %s\nBody: %s\nMessage ID: %s\n",
        to, subject, body, *result.MessageId)
    return nil
}
