# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""Main python script for calculating OSS Criticality Score."""

import argparse
import csv
import datetime
import json
import math
import os
import sys
import time
import urllib

import random
import requests
import csv

languages = set(['python', 'java', 'rust', 'go', 'cplusplus'])
base_url = 'https://storage.googleapis.com/ossf-criticality-score/{lang}_top_200.csv'

def main():
    parser = argparse.ArgumentParser(
        description='Starts the criticality game! Thanks for helping :)')
    args = parser.parse_args()
    play(args)

def get_language():
    while True:
        language=input('Which language would you like to help rank? Choices are: {0}\n'.format(', '.join(languages)))
        if language not in languages:
            print('Invalid Language. Try again!')
        return language

results=[]

def play(args):
    print('Thanks for playing the criticality game! When you finish, please consider sharing your results')
    lang=get_language()
    url=base_url.format(lang=lang)
    top_list=requests.get(url)
    dep_names=[d.split(',')[0] for d in top_list.text.splitlines()[1:]]
    
    # Main game loop!
    print('Exit at any time with ctrl+c or typing done!')
    while True:
        choice1=random.choice(dep_names)
        choice2=random.choice(dep_names)
        if choice1 == choice2:
                continue
        result=pick_winner(choice1, choice2)
        results.append(result)
    

def pick_winner(c1, c2):
    while True:
        winner=input('Which library do you think is more **critical**? Type [0] for {0}, [1] for {1} or [done] to exit.\n'.format(c1, c2))
        if winner == 'done':
            cleanup()
        if winner == '0' or winner == c1:
            return c1, c2
        elif winner == '1' or c2:
            return c2, c1
        print('Oops, invalid answer. Try again.')


def cleanup():
    print('Thanks for playing! Here are your results. Please consider sharing them with us!')
    for result in results:
        print(' '.join(result))
    exit(0)

        
            


        

if __name__ == "__main__":
    main()