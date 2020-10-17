package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/ogrok/crmd/models"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	Directory   = "/.crmd"
	StorageFile = "/reminders.json"

	RecurrenceDaily = "daily"
	RecurrenceWeekly = "weekly"
	RecurrenceMonthly = "monthly"
	RecurrenceQuarterly = "quarterly"
	RecurrenceYearly = "yearly"
)

func main() {
	// first init array of valid recurrences
	validRecurrences := []string{
		RecurrenceDaily,
		RecurrenceWeekly,
		RecurrenceMonthly,
		RecurrenceQuarterly,
		RecurrenceYearly,
	}

	// list possible flags to be inserted into here
	var opts struct {
		Date     string `short:"d" description:"Date of new reminder"`
		Time     string `short:"t" description:"Optional time of new reminder"`
		Recur    string `short:"r" description:"Recurrence schedule for new reminder"`
		Complete int    `short:"c" description:"ID of reminder to mark complete"`
		Delete   int    `short:"x" description:"ID of reminder to delete"`
	}

	descArray, _ := flags.Parse(&opts)
	description := strings.Join(descArray, " ")

	hasDate := strPopulated(opts.Date)
	hasTime := strPopulated(opts.Time)
	hasRecur := strPopulated(opts.Recur)
	hasDescription := slicePopulated(descArray)
	hasComplete := isNonzero(opts.Complete)
	hasDelete := isNonzero(opts.Delete)

	// if recurrence is passed, validate it
	if hasRecur {
		validRecur := false
		for _, v := range validRecurrences {
			if v == opts.Recur {
				validRecur = true
				break
			}
		}

		if !validRecur {
			fmt.Println("invalid recurrence schedule: "+opts.Recur+"\n"+
				"valid schedules: daily, weekly, monthly, quarterly, yearly")
			os.Exit(0)
		}
	}

	// a couple commands are only valid with no extraneous things
	validComplete := true
	if hasDate || hasTime || hasRecur || hasDescription || hasDelete {
		validComplete = false
	}
	validDelete := true
	if hasDate || hasTime || hasRecur || hasDescription || hasComplete {
		validDelete = false
	}

	if hasComplete {
		if !validComplete {
			fmt.Println("can only use -c flag by itself")
			os.Exit(0)
		}

		result, err := completeReminder(opts.Complete, true)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println(result)
		os.Exit(0)
	}

	if hasDelete {
		if !validDelete {
			fmt.Println("can only use -x flag by itself")
			os.Exit(0)
		}

		result, err := completeReminder(opts.Delete, false)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println(result)
		os.Exit(0)
	}

	if hasDescription {
		if !hasDate {
			fmt.Println("cannot create reminder: no date provided")
			os.Exit(0)
		}

		timestamp, err := toUnixDate(opts.Date, opts.Time)
		if err != nil {
			 fmt.Println(err.Error())
		}

		result, err := createReminder(description, timestamp, opts.Recur)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println(result)
		os.Exit(0)
	} else {
		if hasDate || hasTime || hasRecur {
			fmt.Println("description not found; ignoring flags")
		}

		checkReminders()
	}
}

// FILE I/O

func createReminder(description string, datetime int64, recurrence string) (string, error) {
	// TODO: Implement createReminder().
	return "", nil
}

func checkReminders() {
	// TODO: Implement checkReminders(). This function should be responsible for own output.
}

func completeReminder(id int, allowRecurrence bool) (string, error) {
	// TODO: Implement completeReminder().
	return "", nil
}

func loadRemindersFile() (map[int]models.Reminder, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	directory := home + Directory
	fileLocation := directory + StorageFile

	if !fileExists(fileLocation) {
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			_ = os.Mkdir(directory, 0777)
		}

		newFile, err := os.Create(fileLocation)
		if err != nil {
			return nil, err
		}

		defer newFile.Close()

		_, err = newFile.WriteString("[]\n")
		if err != nil {
			return nil, err
		}
	}

	jsonFile, err := os.Open(fileLocation)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	var reminders []models.Reminder
	json.Unmarshal(bytes, &reminders)

	output := map[int]models.Reminder{}
	for _, v := range reminders {
		output[v.ID] = v
	}

	return output, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// HELPERS

func toUnixDate(d string, t string) (int64, error) {
	// try to parse with both date and time
	result, err := time.ParseInLocation("2006-01-02 15:04", d+" "+t, time.Local)

	if err != nil {
		if len(t) > 0 {
			fmt.Println("note: could not parse time component: "+t)
		}
		result, err = time.ParseInLocation("2006-01-02", d, time.Local)
		if err != nil {
			return 0, errors.New("could not parse time: "+ d+" "+t)
		}
	}

	return result.Unix(), nil
}

func strPopulated(input string) bool {
	return len(input) > 0
}

func slicePopulated(input []string) bool {
	return len(input) > 0
}

func isNonzero(input int) bool {
	return input != 0
}
