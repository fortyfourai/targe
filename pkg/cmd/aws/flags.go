package aws

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func RegisterUsersFlags(flags *pflag.FlagSet) {
	var err error
	if err = viper.BindPFlag("user", flags.Lookup("user")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("operation", flags.Lookup("operation")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("group", flags.Lookup("group")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("policy", flags.Lookup("policy")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("resource", flags.Lookup("resource")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("service", flags.Lookup("service")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("policy_option", flags.Lookup("policy-option")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("terraform", flags.Lookup("terraform")); err != nil {
		panic(err)
	}
}

func RegisterRolesFlags(flags *pflag.FlagSet) {
	var err error
	if err = viper.BindPFlag("role", flags.Lookup("role")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("operation", flags.Lookup("operation")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("policy", flags.Lookup("policy")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("resource", flags.Lookup("resource")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("service", flags.Lookup("service")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("policy_option", flags.Lookup("policy-option")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("terraform", flags.Lookup("terraform")); err != nil {
		panic(err)
	}
}

func RegisterGroupsFlags(flags *pflag.FlagSet) {
	var err error
	if err = viper.BindPFlag("group", flags.Lookup("group")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("operation", flags.Lookup("operation")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("policy", flags.Lookup("policy")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("resource", flags.Lookup("resource")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("service", flags.Lookup("service")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("policy_option", flags.Lookup("policy-option")); err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("terraform", flags.Lookup("terraform")); err != nil {
		panic(err)
	}
}
