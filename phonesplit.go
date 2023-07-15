package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"strconv"

	"github.com/anvari1313/splitwise.go"
)

const (
	credentialsFile = "credentialss.json"
	tokenFile       = "token.json"
)

func main() {
	// Read credentials file
	credentials, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	// Parse credentials
	config, err := google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse credentials: %v", err)
	}

	// Check if token file exists
	_, err = os.Stat(tokenFile)
	tokenExists := !os.IsNotExist(err)

	// If token file exists, load the token
	var token *oauth2.Token
	if tokenExists {
		token, err = loadTokenFromFile()
		if err != nil {
			log.Fatalf("Unable to load token from file: %v", err)
		}
	} else {
		// If token file doesn't exist, perform the authorization flow
		token, err = getTokenFromWeb(config)
		if err != nil {
			log.Fatalf("Unable to retrieve token from web: %v", err)
		}

		// Save token to file
		err = saveTokenToFile(token)
		if err != nil {
			log.Fatalf("Unable to save token to file: %v", err)
		}
	}

	// Create Gmail service client
	client := config.Client(context.Background(), token)
	service, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create Gmail service: %v", err)
	}

	// Call Gmail API
	// Example: List all messages in the user's mailbox

	query := "from:noreply@phonepe.com"

	// get the previous 1 month of the mails
	// Calculate date range for previous month
	now := time.Now()
	start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	// Format date range in query format (yyyy/mm/dd)
	dateQuery := fmt.Sprintf("after:%s before:%s", start.Format("2006/01/02"), end.Format("2006/01/02"))


	user := "me"
	r, err := service.Users.Messages.List(user).Q(query).Q(dateQuery).Do()
	r, err = service.Users.Messages.List(user).Q(query).MaxResults(3).Do()
	
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	if len(r.Messages) == 0 {
		fmt.Println("No messages found.")
	}

	messages := r.Messages

// Loop through the retrieved messages and extract the amount and message from PhonePe emails
	for _, message := range messages {
		msg, err := service.Users.Messages.Get("me", message.Id).Do()
		if err != nil {
			log.Printf("Failed to retrieve message with ID %s: %v", message.Id, err)
			continue
		}

	payload := msg.Payload
	headers := payload.Headers
	var msgBody string
	var amount string
	var transactionID string
	email := getFirstName(*service)
	for _, header := range headers {
		if header.Name == "Subject" && strings.Contains(header.Value, "Sent") {
			parts := payload.Parts
			for _, part := range parts {
				if part.MimeType == "text/plain" {
					data, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err != nil {
						log.Printf("Failed to decode message body: %v", err)
						continue
					}
					sdata2, err := ioutil.ReadAll(bytes.NewReader(data))
					if err != nil {
						log.Printf("Failed to read decoded message body: %v", err)
						continue
					}
					sdataStr := string(sdata2)
					sdataStr = strings.ReplaceAll(sdataStr, " ", "")

					sdata := sdataStr

					msgBody = sdata[strings.Index(sdata, "Message:")+len("Message")+1 : strings.Index(sdata, "Hi")]
					amountStart := strings.Index(sdata, "₹") + len("₹")
					amountEnd := strings.Index(sdata, "Txn")
					if amountStart != -1 && amountEnd != -1 {
						amount = sdata[amountStart:amountEnd]
					}
					transactionID = sdata[strings.Index(sdata, "Txn.ID:")+len("Txn.ID:") : strings.Index(sdata, "Txn.status")]
					fmt.Println("Message", msgBody)
					fmt.Println("Amount", amount)
					fmt.Println("Transaction ID", transactionID)
					
					if msgBody != "" {
						result := true
						for _, char := range msgBody {
							if char!='p' && char != 'k' && char != 'y' && char != 's'{
								result = false
							}
						}
						if result {
							err := AddToSplitWise(email,msgBody,amount,transactionID)
							if err != nil{
								panic(err)
							}
						} else {
							fmt.Println("Invalid Message Found.")
						}
					}
				}
			}
		}
	}
	
	fmt.Println("-----------")
}

}

func getFirstName(service gmail.Service) string {
	user := "me"
	profile, err := service.Users.GetProfile(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve user profile: %v", err)
	}

	emailParts := strings.Split(profile.EmailAddress, "@")
	firstName := strings.Split(emailParts[0], ".")[0]
	return firstName
}

func loadTokenFromFile() (*oauth2.Token, error) {
	file, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func saveTokenToFile(token *oauth2.Token) error {
	file, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(token)
	if err != nil {
		return err
	}

	return nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}


func AddToSplitWise(emailInitial string , msg string, amountPaid string, transaction_id string) error {

	// amount to int
	amountPaidInt, err := strconv.Atoi(amountPaid)
	if err != nil {
		fmt.Printf("Failed to convert string to int: %v\n", err)
		return err
	}

  auth := splitwise.NewAPIKeyAuth("TeivHyZqXSI5LfkEGkw33XY2gYglpC28teQcQCz8")
  client := splitwise.NewClient(auth)

	//groups by id
	group, err := client.GroupByID(context.Background(), 50986733)
	if err != nil {
		panic(err)
	}

	isPresent,err := checkExpenses(client,transaction_id)
	if err != nil{
		return err
	}

	if !isPresent{

		var botUser []int
		var master splitwise.GroupMember

		msg = strings.ToLower(msg)
		for _,mem := range group.Members{
			// email needs to start with the firstname initial
			if strings.ToLower(string(mem.FirstName[0]))== strings.ToLower(string(emailInitial[0])) {
				master = mem
			}
			firstName := strings.ToLower(mem.FirstName)
			char := rune(firstName[0]) // Convert the byte to rune
			if !isCharPresent(msg,char){
					botUser = append(botUser, mem.ID)
					continue
			}
		}

		var isMasterIncludedInTran = true
		// additional condition check for master
		firstName := strings.ToLower(master.FirstName)
		char := rune(firstName[0])
		if !isCharPresent(msg,rune(char)){
			isMasterIncludedInTran = false 
		}
		// temp add Example
		err = addExpense(client, botUser, master,group.Members, amountPaidInt,50986733,transaction_id, isMasterIncludedInTran)
		if err != nil {
		panic(err)
		}
	} else {
		fmt.Println("Transaction ID already Present")
	}
	 return nil
	
}


func addExpense(client splitwise.Client,botUser  []int, master splitwise.GroupMember, slaves []splitwise.GroupMember, amountPaid int, groupID int, 
	transactionID string, isMasterIncluded bool) error {
	// Prepare the expense data
	expense := splitwise.Expense{
			Cost:        fmt.Sprintf("%.2f", float64(amountPaid)),
			Description: fmt.Sprintf("Auto Added By TerminalKar. Transaction ID:%s",transactionID),
			GroupId:     uint32(groupID),
			CurrencyCode: "IRR",
	}

	// Create the user shares
	var usersShares []splitwise.UserShare
	for _, slave := range slaves {

		if slave.ID == master.ID {
			var owedShare = fmt.Sprintf("%.2f", float64(amountPaid)/float64(len(slaves)-len(botUser)))
			if !isMasterIncluded{
				owedShare = fmt.Sprintf("0")					
			}
			userShares := splitwise.UserShare{
				UserID:    uint64(slave.ID),
				PaidShare: fmt.Sprintf("%.2f",float64(amountPaid)),
				OwedShare: owedShare,
			}
			usersShares = append(usersShares, userShares)

		} else if isIntPresent(botUser,slave.ID) {
			userShares := splitwise.UserShare{
				UserID:    uint64(slave.ID),
				PaidShare: "0",
				OwedShare: "0",
			}
			usersShares = append(usersShares, userShares)
		} else {
			userShares := splitwise.UserShare{
				UserID:    uint64(slave.ID),
				PaidShare: "0.00",
				OwedShare: fmt.Sprintf("%.2f", float64(amountPaid)/float64(len(slaves)-len(botUser))),
			}
			usersShares = append(usersShares, userShares)
		}
	}

	// Call the Splitwise API to create the expense
	resp, err := client.CreateExpenseByShare(context.Background(),expense, usersShares)
	if err != nil {
		fmt.Println("Error creating expense:", err)
		return err
	}

	if len(resp) == 0 {
		fmt.Println("There is some issue with the splitwise api. Maybe a Logic Issue from our side.")
		return nil
	}

	fmt.Println("Transaction Successfully Added To Splitwise.")

	return nil
}


func checkExpenses(client splitwise.Client, transaction_id string) (bool,error) {
	expensesRes, err := client.Expenses(context.Background())
	if err != nil {
		panic(err)
	}

	if len(expensesRes) != 0 {
		for _, expense := range expensesRes {
			if strings.Contains(expense.Description, transaction_id) && expense.DeletedAt == ""{
				return true,nil
			}
		}
	}
	return false,nil

}

func isCharPresent(str string, char rune) bool {
    for _, c := range str {
        if c == char {
            return true
        }
    }
    return false
}

func isIntPresent(arr []int, target int) bool {
    for _, num := range arr {
        if num == target {
            return true
        }
    }
    return false
}